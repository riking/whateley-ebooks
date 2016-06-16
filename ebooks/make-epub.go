package ebooks

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"time"
	"os"

	"golang.org/x/net/html"
	"github.com/pkg/errors"
	"io/ioutil"
	"sync"
	"net/http"
	"github.com/riking/whateley-ebooks/client"
	"bufio"
	"github.com/andybalholm/cascadia"
)

// EpubDefinition is the setup, layout, and metadata of a target .epub file.
// It also contains the working state of an epub being created.
// EpubDefinition is not safe for multithreaded use - see Clone().
type EpubDefinition struct {
	Parts []struct {
		// Table of Contents entry
		TOC       string
		CoverPage string
		Story     struct {
			ID           string
			Slug         string
			RemoveTitles string `yaml:"remove-titles"`
		}
	}
	Assets []struct {
		// the URL to download the asset from (should start with http://whateleyacademy.net)
		Download string
		// the location in the epub of the asset
		Target string
		// content to search 'src' attributes for
		Find string
		// what to replace matching 'src' attributes with
		Replace string
	}

	Author     string
	AuthorSort string `yaml:"author-sort"`
	Title      string
	TitleSort  string `yaml:"title-sort"`
	Publisher  string
	Series     string
	UUID       string

	files contentEntries
	workingDir string
	lock sync.Mutex
}

func (ed *EpubDefinition) Clone() *EpubDefinition {
	return &EpubDefinition{
		Parts: ed.Parts,
		Assets: ed.Assets,
		Author: ed.Author,
		AuthorSort: ed.AuthorSort,
		Title: ed.Title,
		TitleSort: ed.TitleSort,
		Publisher: ed.Publisher,
		Series: ed.Series,
		UUID: ed.UUID,
	}
}

func (ed *EpubDefinition) AuthorFileAs() string {
	if ed.AuthorSort == "" {
		return ed.Author
	}
	return ed.AuthorSort
}

func (ed *EpubDefinition) Date() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05-07:00")
}

func (ed *EpubDefinition) TitleFileAs() string {
	if ed.TitleSort == "" {
		return ed.Title
	}
	return ed.TitleSort
}

type contentEntry struct {
	Filename    string
	Id          string
	ContentType string
	TOC         string
}

func (c *contentEntry) RenderManifest(w io.Writer) {
	fmt.Fprintf(w, `<item href="%s" id="%s" media-type="%s"/>`,
		html.EscapeString(c.Filename), html.EscapeString(c.Id), html.EscapeString(c.ContentType))
}

func (c *contentEntry) RenderSpine(w io.Writer) {
	if c.ContentType == "application/xhtml+xml" {
		fmt.Fprintf(w, `<itemref idref="%s"/>`, html.EscapeString(c.Id))
	}
}

type contentEntries []contentEntry

func (c contentEntries) RenderInContentOPF(w io.Writer) {
	fmt.Fprintf(w, "<manifest>")
	for _, v := range c {
		v.RenderManifest(w)
	}
	fmt.Fprintf(w, `</manifest><spine toc="ncx">`)
	for _, v := range c {
		v.RenderSpine(w)
	}
	fmt.Fprintf(w, `</spine>`)
}

func (c contentEntries) RenderInTocNCX(w io.Writer) {
	// TODO
}

func (ed *EpubDefinition) WriteStyles() error {
	ebookDir := fmt.Sprintf("%s/OEBPS", ed.workingDir)
	os.MkdirAll(fmt.Sprintf("%s/Styles", ebookDir), 0755)

	id := "story.css"
	filename := "Styles/story.css"
	file, err := os.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
	if err != nil {
		return errors.Wrapf(err, "creating target file for asset %s", id)
	}
	b, err := Asset("story.css")
	if err != nil {
		return errors.Wrapf(err, "could not find embedded asset %s", id)
	}
	w := bufio.NewWriter(file)
	_, err = w.Write(b)
	if err != nil {
		return errors.Wrapf(err, "writing asset file %s", id)
	}
	err = w.Flush()
	if err != nil {
		return errors.Wrapf(err, "writing asset file %s", id)
	}
	err = file.Close()
	if err != nil {
		return errors.Wrapf(err, "writing asset file %s", id)
	}
	return nil
}

var contentOPFTmpl = template.Must(template.New("content.opf").Parse(string(MustAsset("content.opf"))))

func (ed *EpubDefinition) ManifestAndSpine() template.HTML {
	var buf bytes.Buffer
	ed.files.RenderInContentOPF(&buf)
	return template.HTML(buf.String())
}

func (ed *EpubDefinition) RenderContentOPF(w io.Writer) error {
	return contentOPFTmpl.Execute(w, ed)
}

func (ed *EpubDefinition) DownloadAssets(access *client.WANetwork) error {
	ebookDir := fmt.Sprintf("%s/OEBPS", ed.workingDir)
	os.MkdirAll(fmt.Sprintf("%s/Images", ebookDir), 0755)
	for _, v := range ed.Assets {
		filename := fmt.Sprintf("Images/%s", v.Target)
		file, err := os.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
		if err != nil {
			return errors.Wrapf(err, "creating target file for asset %s", v.Target)
		}

		req, err := http.NewRequest("GET", v.Download, nil)
		if err != nil {
			panic(err)
		}

		resp, err := access.Do(req)
		if err != nil {
			return errors.Wrapf(err, "downloading asset %s", v.Download)
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return errors.Wrapf(err, "writing asset file %s", v.Target)
		}
		err = resp.Body.Close()
		if err != nil {
			return errors.Wrapf(err, "writing asset file %s", v.Target)
		}
		err = file.Close()
		if err != nil {
			return errors.Wrapf(err, "writing asset file %s", v.Target)
		}

		ed.files = append(ed.files, contentEntry{
			Filename: filename,
			Id: v.Target,
			ContentType: resp.Header.Get("Content-Type"),
		})
	}
	return nil
}

var coverPageTmpl = template.Must(template.New("cover-page").Parse(string(MustAsset("cover.html"))))

var storyPageTmpl = template.Must(template.New("story-page").Parse(string(MustAsset("part.html"))))

func findMatchingSrc(src string) func(*html.Node) bool {
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		for _, v := range n.Attr {
			if v.Key == "src" {
				return v.Val == src
			}
		}
		return false
	}
}

func (ed *EpubDefinition) CreateTextPages(access *client.WANetwork) error {
	ebookDir := fmt.Sprintf("%s/OEBPS", ed.workingDir)
	os.MkdirAll(fmt.Sprintf("%s/Text", ebookDir), 0755)

	coverCount := 0
	for _, v := range ed.Parts {
		var filename string
		if v.CoverPage != "" {
			coverCount++
			id := fmt.Sprintf("Cover%02d.html", coverCount)
			filename = fmt.Sprintf("Text/%s", id)
			file, err := os.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
			if err != nil {
				return errors.Wrapf(err, "creating target file for %s", filename)
			}
			w := bufio.NewWriter(file)
			w.WriteString(`<?xml version="1.0" encoding="utf-8" standalone="no"?>`)
			err = coverPageTmpl.Execute(w, template.HTML(v.CoverPage))
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}
			err = w.Flush()
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}
			err = file.Close()
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}

			ed.files = append(ed.files, contentEntry{
				Filename: filename,
				Id: id,
				ContentType: "application/xhtml+xml",
				TOC: v.TOC,
			})
		} else {
			page, err := access.GetStoryByID(v.Story.ID)
			if err != nil {
				return errors.Wrapf(err, "getting story %s (%s)", v.Story.ID, v.Story.Slug)
			}
			id := fmt.Sprintf("%s-%s.html", page.StoryID, page.StorySlug)
			filename = fmt.Sprintf("Text/%s", id)
			file, err := os.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
			if err != nil {
				return errors.Wrapf(err, "creating target file for %s", filename)
			}

			err = FixForEbook(page)
			if err != nil {
				return errors.Wrapf(err, "preparing story %s", page.StorySlug)
			}
			// perform RemoveTitles
			page.Doc().Find(v.Story.RemoveTitles).Remove()
			// perform asset replacement
			for _, asset := range ed.Assets {
				page.StoryBodySelection().FindMatcher(cascadia.Selector(findMatchingSrc(asset.Find))).SetAttr("src", asset.Replace)
			}

			w := bufio.NewWriter(file)
			w.WriteString(`<?xml version="1.0" encoding="utf-8" standalone="no"?>`)
			err = storyPageTmpl.Execute(w, page)
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}
			err = w.Flush()
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}
			err = file.Close()
			if err != nil {
				return errors.Wrapf(err, "writing target file for %s", filename)
			}

			ed.files = append(ed.files, contentEntry{
				Filename: filename,
				Id: id,
				ContentType: "application/xhtml+xml",
				TOC: v.TOC,
			})
		}
	}
	return nil
}

func CreateEpub(ed *EpubDefinition, access *client.WANetwork, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}
	defer file.Close()

	workDir, err := ioutil.TempDir("target", "epub-tmp")
	if err != nil {
		return errors.Wrap(err, "could not create tmpdir")
	}
	defer os.RemoveAll(workDir)

	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.workingDir = workDir

	// Download assets
	err = ed.DownloadAssets(access)
	if err != nil {
		return err
	}

	err = ed.CreateTextPages(access)
	if err != nil {
		return err
	}

	err = ed.WriteStyles()
	if err != nil {
		return err
	}

	err = ed.RenderContentOPF(file)
	if err != nil {
		return err
	}

	fmt.Println(ed.files)

	os.Exit(3)
	return nil
}
