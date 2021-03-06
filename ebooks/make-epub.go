// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package ebooks

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"sync"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/denisbrodbeck/machineid"
	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/riking/whateley-ebooks/client"
)

// A TOCEntry is one part of the backbone of the epub file.
type TOCEntry struct {
	// Table of Contents entry
	TOC       string
	TOCNest   int    `yaml:"toc-nest"`
	TOCPage   string `yaml:"toc-page"`
	CoverPage string `yaml:"coverpage"`
	fixCvr    bool   `yaml:"-"`
	Story     struct {
		ID           string
		Slug         string
		RemoveTitles string `yaml:"remove-titles"`
		page         *client.WhateleyPage
	}
}

func (t *TOCEntry) IsCoverPage() bool {
	return t.CoverPage != ""
}

func (t *TOCEntry) IsContentPage() bool {
	return t.Story.ID != ""
}

func (t *TOCEntry) IsAnchorEntry() bool {
	return t.TOCPage != ""
}

// called from CreateEpub, so short-circuit
func (t *TOCEntry) preparePage(access *client.WANetwork, ed *EpubDefinition) (*client.WhateleyPage, error) {
	if t.Story.page != nil {
		return t.Story.page, nil
	}
	if !t.IsContentPage() {
		return nil, errors.Errorf("Not a content page")
	}

	page, err := t.preparePageA(access, ed)
	if err != nil {
		return nil, err
	}
	err = t.preparePageB(access, ed, page)
	if err != nil {
		return nil, err
	}
	return t.Story.page, nil
}

// called from EpubDefinition.Prepare()
func (t *TOCEntry) preparePageA(access *client.WANetwork, ed *EpubDefinition) (*client.WhateleyPage, error) {
	if !t.IsContentPage() {
		return nil, errors.Errorf("Not a content page")
	}

	page, err := access.GetStoryByID(t.Story.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "getting story %s (%s)", t.Story.ID, t.Story.Slug)
	}

	return page, nil
}

// preparePageB performs the processing on a downloaded content page and stores the result in an internal field.
// The page parameter is mutated.
func (t *TOCEntry) preparePageB(access *client.WANetwork, ed *EpubDefinition, page *client.WhateleyPage) error {
	if t.Story.page != nil {
		return nil
	}
	if !t.IsContentPage() {
		return errors.Errorf("Not a content page")
	}

	err := FixForEbook(page)
	if err != nil {
		return errors.Wrapf(err, "preparing story %s", page.StorySlug)
	}
	// perform RemoveTitles
	page.Doc().Find(t.Story.RemoveTitles).Remove()
	// perform asset replacement
	for _, asset := range ed.Assets {
		s := page.StoryBodySelection().FindMatcher(cascadia.Selector(findMatchingSrc(asset.Find)))
		if s.Length() > 0 {
			s.SetAttr("src", fmt.Sprintf("../Images/%s", asset.Target))
		}
	}

	t.Story.page = page
	return nil
}

// EpubDefinition is the setup, layout, and metadata of a target .epub file.
// It also contains the working state of an epub being created.
// EpubDefinition is not safe for multithreaded use - see Clone().
type EpubDefinition struct {
	Parts  []TOCEntry
	Assets []struct {
		// the URL to download the asset from (should start with http://whateleyacademy.net)
		Download string
		// the location in the epub of the asset
		Target string
		// content to search 'src' attributes for
		Find string
	}

	Author     string
	AuthorSort string `yaml:"author-sort"`
	Title      string
	TitleSort  string `yaml:"title-sort"`
	Publisher  string
	Series     string
	UUID       string

	files     contentEntries
	lock      sync.Mutex
	wordCount int

	assetPrepareDone bool
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

func (ed *EpubDefinition) Hostname() (string, error) {
	return os.Hostname()
}

func (ed *EpubDefinition) MachineID() (string, error) {
	return machineid.ID()
}

func (ed *EpubDefinition) TitleFileAs() string {
	if ed.TitleSort == "" {
		return ed.Title
	}
	return ed.TitleSort
}

func (ed *EpubDefinition) Username() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "<ERROR>", err
	}
	return u.Name, nil
}

func (ed *EpubDefinition) PrepareAssets() {
	if ed.assetPrepareDone {
		return
	}
	for i := range ed.Assets {
		v := &ed.Assets[i]
		if v.Download == "" {
			v.Download = fmt.Sprintf("http://whateleyacademy.net/images/%s", v.Target)
		}
		if v.Find == "" {
			v.Find = fmt.Sprintf("/images/%s", v.Target)
		}
	}
	ed.Parts = append([]TOCEntry{{fixCvr: true}}, ed.Parts...)
	ed.assetPrepareDone = true
}

func (ed *EpubDefinition) Prepare(access *client.WANetwork) error {
	ed.lock.Lock()
	defer ed.lock.Unlock()

	ed.PrepareAssets()

	const netParallel = 10
	var procParallel = runtime.NumCPU()
	type result error
	type downloaded struct {
		t *TOCEntry
		p *client.WhateleyPage
	}

	tocEntryChan := make(chan *TOCEntry)
	var networkWG sync.WaitGroup
	downloadedChan := make(chan downloaded)
	var processWG sync.WaitGroup
	resultChan := make(chan result)

	// Generator
	go func() {
		for i := range ed.Parts {
			if ed.Parts[i].IsContentPage() {
				tocEntryChan <- &ed.Parts[i]
			}
		}
		close(tocEntryChan)
	}()

	// Network Workers
	networkWG.Add(netParallel)
	for i := 0; i < netParallel; i++ {
		go func() {
			for v := range tocEntryChan {
				p, err := v.preparePageA(access, ed)
				if err != nil {
					resultChan <- err
				} else {
					downloadedChan <- downloaded{t: v, p: p}
				}
			}
			networkWG.Done()
		}()
	}

	// Closer
	go func() {
		networkWG.Wait()
		close(downloadedChan)
	}()

	// CPU Workers
	processWG.Add(procParallel)
	for i := 0; i < procParallel; i++ {
		go func() {
			for v := range downloadedChan {
				err := v.t.preparePageB(access, ed, v.p)
				resultChan <- errors.Wrapf(err, "Story %s failed", v.t.Story.ID)
			}
			processWG.Done()
		}()
	}

	// Closer
	go func() {
		processWG.Wait()
		close(resultChan)
	}()

	// Collector
	errs := []error{}
	for v := range resultChan {
		err := error(v)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 1 {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "Failed to prepare story parts:\n")
		fmt.Fprintf(&buf, "%+v\n", errs[0])

		for _, v := range errs[1:] {
			fmt.Fprintf(&buf, "%s\n", v)
		}
		return errors.New(buf.String())
	} else if len(errs) == 1 {
		return errors.Wrap(errs[0], "Failed to prepare story parts")
	} else {
		return nil
	}
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
	fmt.Fprintf(w, `<navPoint id="navPoint-%d" playOrder="%d"><navLabel><text>%s</text></navLabel><content src="%s"/>`,
		sequence, sequence, template.HTMLEscapeString(c.TOC), template.HTMLEscapeString(c.Filename))
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
<!-- epub file generated by https://github.com/riking/whateley-ebooks -->
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
<!-- epub file generated by https://github.com/riking/whateley-ebooks -->
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

		if v.Download == "" {
			v.Download = fmt.Sprintf("http://whateleyacademy.net/images/%s", v.Target)
		}

		req, err := http.NewRequest("GET", v.Download, nil)
		if err != nil {
			panic(err)
		}

		body, contentType, err := access.GetAsset(req)
		if err != nil {
			return errors.Wrapf(err, "downloading asset %s", v.Download)
		}

		_, err = file.Write(body)
		if err != nil {
			return errors.Wrapf(err, "writing asset file %s", v.Target)
		}

		ed.files = append(ed.files, contentEntry{
			Filename:    filename,
			Id:          v.Target,
			ContentType: contentType,
		})
	}
	return nil
}

var aboutPageTmpl = template.Must(template.New("about-page").Parse(string(MustAsset("about.html"))))
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
		if v.IsCoverPage() && !v.fixCvr {
			coverCount++
			id := fmt.Sprintf("Cover%02d.html", coverCount)
			filename = "Text/" + id
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
		} else if v.fixCvr {
			id := "About.html"
			filename := "Text/" + id
			file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
			if err != nil {
				return errors.Wrapf(err, "creating target file for %s", filename)
			}
			w := bufio.NewWriter(file)
			w.WriteString(`<?xml version="1.0" encoding="utf-8" standalone="no"?>`)
			err = aboutPageTmpl.Execute(w, ed)
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
				TOC:         "",
				TOCNest:     0,
			})
		} else if v.IsContentPage() {
			page, err := v.preparePage(access, ed)
			if err != nil {
				return errors.Wrapf(err, "preparing content for story #%s", v.Story.ID)
			}
			id := fmt.Sprintf("%s-%s.html", page.StoryID, page.StorySlug)
			filename := "Text/" + id
			file, err := fs.Create(fmt.Sprintf("%s/%s", ebookDir, filename))
			if err != nil {
				return errors.Wrapf(err, "creating target file for %s", filename)
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
		} else if v.IsAnchorEntry() {
			// TOC entry to an anchor on existing page
			ed.files = append(ed.files, contentEntry{
				Filename: v.TOCPage,
				TOC:      v.TOC,
				TOCNest:  v.TOCNest,
			})
		} else {
			return errors.Errorf("bad epub definition file [%#v]", v)
		}
	}
	return nil
}

type fileCreator interface {
	Create(name string) (io.Writer, error)
}

type panicDupesZipWriter struct {
	*zip.Writer
	files map[string]struct{}
}

func (w panicDupesZipWriter) Create(name string) (io.Writer, error) {
	if _, ok := w.files[name]; ok {
		return nil, errors.Errorf("Attempt to create duplicate file name %s", name)
	}
	w.files[name] = struct{}{}
	return w.Writer.Create(name)
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

	ed.PrepareAssets()

	_zipWriter := zip.NewWriter(file)
	zipWriter := panicDupesZipWriter{
		_zipWriter,
		make(map[string]struct{}),
	}

	err = ed.WriteMetaINF(zipWriter)
	if err != nil {
		return err
	}

	ed.lock.Lock()
	defer ed.lock.Unlock()

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
