package ebooks

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"time"

	"bufio"
	"net/http"
	"sync"

	"archive/zip"

	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"github.com/riking/whateley-ebooks/client"
	"golang.org/x/net/html"
)

// EpubDefinition is the setup, layout, and metadata of a target .epub file.
// It also contains the working state of an epub being created.
// EpubDefinition is not safe for multithreaded use - see Clone().
type EpubDefinition struct {
	Parts []struct {
		// Table of Contents entry
		TOC       string
		TOCNest   int    `yaml:"toc-nest"`
		TOCPage   string `yaml:"toc-page"`
		CoverPage string `yaml:"coverpage"`
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

	files      contentEntries
	workingDir string
	lock       sync.Mutex
	wordCount  int
}

func (ed *EpubDefinition) Clone() *EpubDefinition {
	return &EpubDefinition{
		Parts:      ed.Parts,
		Assets:     ed.Assets,
		Author:     ed.Author,
		AuthorSort: ed.AuthorSort,
		Title:      ed.Title,
		TitleSort:  ed.TitleSort,
		Publisher:  ed.Publisher,
		Series:     ed.Series,
		UUID:       ed.UUID,
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
	TOCNest     int
}

func (c *contentEntry) RenderManifest(w io.Writer) {
	if c.Id != "" {
		fmt.Fprintf(w, `<item href="%s" id="%s" media-type="%s"/>`,
			html.EscapeString(c.Filename), html.EscapeString(c.Id), html.EscapeString(c.ContentType))
	}
}

func (c *contentEntry) RenderSpine(w io.Writer) {
	if c.ContentType == "application/xhtml+xml" {
		fmt.Fprintf(w, `<itemref idref="%s"/>`, html.EscapeString(c.Id))
	}
}

// RenderNavPoint returns the new value of nest.
func (c *contentEntry) RenderNavPoint(w io.Writer, sequence, nest int) int {
	var newNest int
	if c.TOCNest == nest {
		fmt.Fprintf(w, "</navPoint>")
		newNest = nest
	} else if c.TOCNest > nest {
		newNest = c.TOCNest
	} else { // c.TOCNest < nest
		newNest = nest
		count := 1
		for c.TOCNest < newNest {
			fmt.Fprintf(w, "</navPoint>")
			newNest--
			count++
		}
		fmt.Fprintf(w, "</navPoint>")
	}
	fmt.Fprintf(w, `
<navPoint id="navPoint-%d" playOrder="%d">
<navLabel><text>%s</text></navLabel>
<content src="%s"/>`, sequence, sequence, template.HTMLEscapeString(c.TOC), template.HTMLEscapeString(c.Filename))
	return newNest
}

type contentEntries []contentEntry

func (c contentEntries) RenderInContentOPF(w io.Writer) {
	fmt.Fprint(w, "<manifest>")
	for _, v := range c {
		v.RenderManifest(w)
	}
	fmt.Fprint(w, `</manifest><spine toc="ncx">`)
	for _, v := range c {
		v.RenderSpine(w)
	}
	fmt.Fprint(w, `</spine>`)
}

// RenderInTocNCX renders the navMap that becomes the table of contents in your ebook reader.
//
// TOC entries can be nested. Here's how.
//
//   - toc: Book 1
//     toc-nest: 1
//   - toc: Chapter 1
//     toc-nest: 2
//   - toc: Chapter 2
//     toc-nest: 2
//   - toc: Book 2
//     toc-nest: 1
func (c contentEntries) RenderInTocNCX(w io.Writer) {
	fmt.Fprint(w, "<navMap>")
	sequence := 0
	nest := 0
	for _, v := range c {
		if v.TOCNest == 0 {
			// default value of 1, since 0 has special meaning
			v.TOCNest = 1
		}
		if v.TOC != "" {
			sequence++
			nest = v.RenderNavPoint(w, sequence, nest)
		}
	}
	count := 0
	for nest > 0 {
		fmt.Fprint(w, "</navPoint>")
		nest--
		count++
	}
	fmt.Fprint(w, "</navMap>")
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

var tocNCXTmpl = template.Must(template.New("toc.ncx").Parse(string(MustAsset("toc.ncx"))))

func (ed *EpubDefinition) NavMap() template.HTML {
	var buf bytes.Buffer
	ed.files.RenderInTocNCX(&buf)
	return template.HTML(buf.String())
}

func (ed *EpubDefinition) RenderTOC(w io.Writer) error {
	return tocNCXTmpl.Execute(w, ed)
}

func (ed *EpubDefinition) WriteTableOfContents(fs fileCreator) error {
	ebookDir := "OEBPS"
	filename := "toc.ncx"
	file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
	if err != nil {
		return errors.Wrapf(err, "creating target file for asset %s", "toc.ncx")
	}
	file.Write([]byte(`<?xml version="1.0" encoding="UTF-8" standalone="no" ?>
<!DOCTYPE ncx PUBLIC "-//NISO//DTD ncx 2005-1//EN"
 "http://www.daisy.org/z3986/2005/ncx-2005-1.dtd">
`))
	err = ed.RenderTOC(file)
	if err != nil {
		return errors.Wrapf(err, "processing asset %s", "toc.ncx")
	}
	ed.files = append(ed.files, contentEntry{
		Filename:    filename,
		Id:          "ncx",
		ContentType: "application/x-dtbncx+xml",
	})
	return nil
}

func (ed *EpubDefinition) WriteContentOPF(fs fileCreator) error {
	ebookDir := "OEBPS"
	filename := "content.opf"

	file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
	if err != nil {
		return errors.Wrapf(err, "creating target file for asset %s", filename)
	}
	file.Write([]byte(`<?xml version="1.0"  encoding="UTF-8"?>
`))
	err = ed.RenderContentOPF(file)
	if err != nil {
		return errors.Wrapf(err, "processing asset %s", filename)
	}
	// this writes out the file list, so no need to add the file to it
	return nil
}

func (ed *EpubDefinition) WriteMetaINF(fs fileCreator) error {
	file, err := fs.Create("mimetype")
	if err != nil {
		return errors.Wrapf(err, "creating target file for asset %s", "mimetype")
	}
	_, err = file.Write([]byte(`application/epub+zip`))
	if err != nil {
		return errors.Wrapf(err, "writing asset file %s", "mimetype")
	}

	//os.MkdirAll(fmt.Sprintf("%s/META-INF", ed.workingDir), 0755)
	file, err = fs.Create("META-INF/container.xml")
	if err != nil {
		return errors.Wrapf(err, "creating target file for asset %s", "META-INF/container.xml")
	}
	_, err = file.Write([]byte(`
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
<rootfiles>
<rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
</rootfiles>
</container>`))
	if err != nil {
		return errors.Wrapf(err, "writing asset file %s", "META-INF/container.xml")
	}
	return nil
}

func (ed *EpubDefinition) WriteStyles(fs fileCreator) error {
	ebookDir := "OEBPS"

	id := "story.css"
	filename := "Styles/story.css"
	file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
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
	ed.files = append(ed.files, contentEntry{
		Filename:    filename,
		Id:          id,
		ContentType: "text/css",
	})
	return nil
}

func (ed *EpubDefinition) WriteAssets(access *client.WANetwork, fs fileCreator) error {
	ebookDir := "OEBPS"
	for _, v := range ed.Assets {
		filename := fmt.Sprintf("Images/%s", v.Target)
		file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
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

		ed.files = append(ed.files, contentEntry{
			Filename:    filename,
			Id:          v.Target,
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

func (ed *EpubDefinition) WriteText(access *client.WANetwork, fs fileCreator) error {
	ebookDir := "OEBPS"

	coverCount := 0
	for _, v := range ed.Parts {
		var filename string
		if v.CoverPage != "" {
			coverCount++
			id := fmt.Sprintf("Cover%02d.html", coverCount)
			filename = fmt.Sprintf("Text/%s", id)
			file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
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

			ed.files = append(ed.files, contentEntry{
				Filename:    filename,
				Id:          id,
				ContentType: "application/xhtml+xml",
				TOC:         v.TOC,
				TOCNest:     v.TOCNest,
			})
		} else if v.Story.ID != "" {
			page, err := access.GetStoryByID(v.Story.ID)
			if err != nil {
				return errors.Wrapf(err, "getting story %s (%s)", v.Story.ID, v.Story.Slug)
			}
			id := fmt.Sprintf("%s-%s.html", page.StoryID, page.StorySlug)
			filename = fmt.Sprintf("Text/%s", id)
			file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
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

			ed.wordCount += page.WordCount()

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

			ed.files = append(ed.files, contentEntry{
				Filename:    filename,
				Id:          id,
				ContentType: "application/xhtml+xml",
				TOC:         v.TOC,
				TOCNest:     v.TOCNest,
			})
		} else if v.TOCPage != "" {
			// TOC entry to an anchor on existing page
			ed.files = append(ed.files, contentEntry{
				Filename: v.TOCPage,
				TOC:      v.TOC,
				TOCNest:  v.TOCNest,
			})
		} else {
			fmt.Println(v)
			panic("bad epub definition file")
		}
	}
	return nil
}

type fileCreator interface {
	Create(name string) (io.Writer, error)
}

type dummyFileCreator func(string) (io.Writer, error)

func (d dummyFileCreator) Create(name string) (io.Writer, error) {
	return d(name)
}

func CreateEpub(ed *EpubDefinition, access *client.WANetwork, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "could not create output file")
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)

	err = ed.WriteMetaINF(zipWriter)
	if err != nil {
		return err
	}

	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.workingDir = "."

	// Download assets
	err = ed.WriteAssets(access, zipWriter)
	if err != nil {
		return err
	}

	err = ed.WriteText(access, zipWriter)
	if err != nil {
		return err
	}

	err = ed.WriteStyles(zipWriter)
	if err != nil {
		return err
	}

	err = ed.WriteTableOfContents(zipWriter)
	if err != nil {
		return err
	}

	err = ed.WriteContentOPF(zipWriter)
	if err != nil {
		return err
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	fmt.Printf("Created %s.\nWord Count: %d\n", filename, ed.wordCount)

	return nil
}
