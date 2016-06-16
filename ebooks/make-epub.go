package ebooks

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"time"

	"golang.org/x/net/html"
)

type EpubDefinition struct {
	Parts []struct {
		// Table of Contents entry
		TOC       string
		CoverPage string
		Story     struct {
			ID           int64
			Slug         string
			RemoveTitles string
		}
	}
	Assets []struct {
		Download string
		Filename string
	}

	Author     string
	AuthorSort string `yaml:"author-sort"`
	Title      string
	TitleSort  string `yaml:"title-sort"`
	Publisher  string
	Series     string
	UUID       string

	files contentEntries
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

var contentOPFTmpl = template.Must(template.New("content.opf").Parse(`
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="BookId" version="2.0">
<metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:opf="http://www.idpf.org/2007/opf">
<dc:creator opf:role="aut" opf:file-as="{{.AuthorFileAs}}">{{.Author}}</dc:creator>
<dc:identifier id="BookId" opf:scheme="UUID">urn:uuid:{{.UUID}}</dc:identifier>
<dc:date>{{.Date}}</dc:date>
<dc:title>{{.Title}}</dc:title>
<dc:publisher>{{.Publisher}}</dc:publisher>
<meta name="calibre:title_sort" content="{{.TitleFileAs}}"/>
<meta name="calibre:series" content="{{.Series}}"/>
<dc:identifier opf:scheme="calibre">{{.UUID}}</dc:identifier>
</metadata>
{{.ManifestAndSpine}}
<guide/>
</package>`))

func (ed *EpubDefinition) ManifestAndSpine() string {
	var buf bytes.Buffer
	ed.files.RenderInContentOPF(&buf)
	return buf.String()
}

func (ed *EpubDefinition) RenderContentOPF(w io.Writer) error {
	return contentOPFTmpl.Execute(w, ed)
}

func CreateEpub(ed *EpubDefinition) {

}
