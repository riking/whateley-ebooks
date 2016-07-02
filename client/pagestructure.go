// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package client // import "github.com/riking/whateley-ebooks/client"

import (
	"fmt"
	"html/template"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

const timeFmt = "2006-01-02T15:04:05-07:00"

type StoryTag struct {
	ID   int
	Name string
}

type StoryURL struct {
	CategorySlug string
	StoryID      string
	StorySlug    string
}

func (u *StoryURL) URL() string {
	return fmt.Sprintf("http://whateleyacademy.net/index.php/%s/%s-%s", u.CategorySlug, u.StoryID, u.StorySlug)
}

func (u *StoryURL) CacheKey() string {
	return fmt.Sprintf("story-%s", u.StoryID)
}

type WhateleyPage struct {
	StoryURL

	document *goquery.Document
	tags     []StoryTag

	// TODO change to funcs
	Previous string
	Next     string
}

func (p *WhateleyPage) Title() string {
	return strings.TrimSpace(p.document.Find(`.item-page .page-header h2[itemprop="name"]`).Text())
}

func (p *WhateleyPage) Authors() string {
	return strings.TrimSpace(p.document.Find(`[itemprop="author"] [itemprop="name"]`).Text())
}

func (p *WhateleyPage) Category() string {
	return strings.TrimSpace(p.document.Find(`.category-name a[itemprop="name"]`).Text())
}

func (p *WhateleyPage) PublishDate() (time.Time, error) {
	t, ok := p.document.Find(`time[itemprop="datePublished"]`).Attr("datetime")
	if !ok {
		return time.Time{}, fmt.Errorf("could not find time.datePublished in %s", p.URL())
	}
	return time.Parse(timeFmt, t)
}

func (p *WhateleyPage) ViewCount() int64 {
	var hits int64
	s, ok := p.document.Find(`.hits [itemprop="interactionCount"]`).Attr("content")
	if !ok {
		fmt.Fprintf(os.Stderr, "<Error: could not find .hits in page %s>\n", p.URL())
		return 0
	}
	fmt.Sscanf(s,
		"UserPageVisits:%d", &hits)
	return hits
}

func (p *WhateleyPage) WordCount() int {
	return len(strings.Fields(p.StoryBodySelection().Text()))
}

const StoryBodySelector = `.item-page div[itemprop="articleBody"] `

func (p *WhateleyPage) StoryBodySelection() *goquery.Selection {
	return p.document.Find(StoryBodySelector)
}

func (p *WhateleyPage) Doc() *goquery.Document {
	return p.document
}

func (p *WhateleyPage) StoryBody() string {
	b, err := p.StoryBodySelection().Html()
	if err != nil {
		panic(err)
	}
	return b
}

func (p *WhateleyPage) StoryBodyForTemplate() template.HTML {
	return template.HTML(p.StoryBody())
}

// TODO - these fail for library, faq, etc
var canonicalURLRegexp = regexp.MustCompile(`\Ahttp://whateleyacademy\.net/index\.php/([a-zA-Z0-9-]+)/(\d+)-([a-zA-Z0-9-]+)`)
var printURLRegexp = regexp.MustCompile(`\A/index.php/(?:([a-zA-Z0-9-]+)/)?(\d+)-([a-zA-Z0-9-]+)(?:(\d+)-([a-zA-Z0-9-]+))?\?tmpl=component&print=1`)

var stripExceptionsSelector = `
head base,
meta[name="rights"],
meta[http-equiv="content-type"],
head title,
div.item-page,
div[itemprop="articleBody"]`

// ParseStoryPage parses a document into a WhateleyPage object.
// Some processing is performed, e.g. elements not relevant are stripped, and the canonical URL is parsed and stored.
func ParseStoryPage(doc *goquery.Document) (*WhateleyPage, error) {
	if doc == nil {
		return nil, errors.Errorf("doc was nil")
	}
	page := new(WhateleyPage)
	doc = goquery.CloneDocument(doc)
	page.document = doc

	// Remove everything not part of the page (header, footer, sidebar)
	// After these two transforms, most of the page bytes will be the actual story
	dontRemove := page.document.Find(stripExceptionsSelector)
	prevLen := 0
	for dontRemove.Length() != prevLen {
		prevLen = dontRemove.Length()
		dontRemove = dontRemove.AddSelection(dontRemove.Children())
	}
	prevLen = 0
	for dontRemove.Length() != prevLen {
		prevLen = dontRemove.Length()
		dontRemove = dontRemove.AddSelection(dontRemove.Parents())
	}

	removed := page.document.Find("html *").NotSelection(dontRemove).RemoveMatcher(cascadia.Selector(func(n *html.Node) bool {
		if n.Type == html.TextNode {
			decision := !dontRemove.IsNodes(n.Parent)
			return decision
		}
		return true
	}))
	_ = removed

	page.document.Find("script,style").Remove()

	// Delete the space nodes adjacent to other space nodes
	spacesTrimmed := 0
	page.document.Find("*").Each(func(_ int, s *goquery.Selection) {
		n := s.Nodes[0]
		if n.Type == html.TextNode && n.NextSibling != nil && n.NextSibling.Type == html.TextNode {
			if strings.TrimSpace(n.Data) == "" && strings.TrimSpace(n.NextSibling.Data) == "" {
				n.NextSibling.Data = "\n"
				s.Remove()
				spacesTrimmed++
				return
			}
		}
		if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" && strings.ContainsRune(n.Data, '\n') {
			n.Data = "\n"
			spacesTrimmed++
			return
		}
	})

	var m []string
	// The printing link is the only part of the page where the correct slug is emitted
	printURL, ok := doc.Find(".print-icon a").Attr("href")
	if ok {
		if strings.Contains(printURL, "the-library") {
			return nil, errors.Errorf("Library stories are not supported at this time (got %s)", printURL)
		}
		// TODO - this fails for library, faq, etc
		m = printURLRegexp.FindStringSubmatch(printURL)
		if m == nil {
			return nil, errors.Errorf("Could not parse canonical URL (got %s)", printURL)
		}
		page.CategorySlug = m[1]
		page.StoryID = m[2]
		page.StorySlug = m[3]
	} else {
		// Fall back on the requested URL
		canonical, ok := doc.Find(`head base`).Attr("href")
		fmt.Fprintf(os.Stderr, "warning: falling back to <base>")
		if !ok {
			return nil, errors.Errorf("could not find <base href> (canonical URL)")
		}
		var err error
		page.StoryURL, err = ParseURL(canonical)
		if err != nil {
			return nil, err
		}
	}

	return page, nil
}

func ParseURL(url string) (StoryURL, error) {
	m := canonicalURLRegexp.FindStringSubmatch(url)
	if m == nil {
		return StoryURL{}, errors.Wrap(fmt.Errorf("got %s", url), "Could not parse canonical story URL")
	}
	return StoryURL{
		CategorySlug: m[1],
		StoryID:      m[2],
		StorySlug:    m[3],
	}, nil
}
