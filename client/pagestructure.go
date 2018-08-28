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
	ID   string
	Slug string
	Name string
}

var tagIdRegexp = regexp.MustCompile(`tag-(\d+)`)

func ParseTag(s *goquery.Selection) StoryTag {
	t := StoryTag{}
	class, _ := s.Attr("class")
	m := tagIdRegexp.FindStringSubmatch(strings.Split(class, " ")[0])
	if m != nil {
		t.ID = m[1]
	}
	l := s.Find("a")
	for _, v := range l.Nodes[0].Attr {
		if v.Key == "href" {
			href := v.Val
			idx := strings.LastIndex(href, "/")
			m = idAndSlugRegexp.FindStringSubmatch(href[idx+1:])
			if m != nil {
				t.Slug = m[2]
			}
		}
	}
	t.Name = strings.TrimSpace(l.Text())
	return t
}

type StoryURL struct {
	CategorySlug string
	StoryID      string
	StorySlug    string
}

func (u *StoryURL) URL() string {
	if u.StoryID == "341" {
		return "http://whateleyacademy.net/index.php/original-timeline/341-tennyo-goes-to-hell-part-1?showall=&start=1"
	} else if u.StoryID == "342" {
		return "http://whateleyacademy.net/index.php/original-timeline/342-tennyo-goes-to-hell-part-2?showall=&start=1"
	}
	return fmt.Sprintf("http://whateleyacademy.net/index.php/original-timeline/%s-%s", u.StoryID, u.StorySlug)
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
	return strings.TrimSpace(p.document.Find(`.category-name a`).Text())
}

func (p *WhateleyPage) CategoryLink() string {
	val, _ := p.document.Find(`.category-name a`).Attr("href")
	return val
}

var serverLoc, _ = time.LoadLocation("America/Los_Angeles")

func (p *WhateleyPage) PublishDate() (time.Time, error) {
	t1 := p.document.Find(`.flexi.element.field_published .value`).Text()
	if t1 != "" {
		return time.ParseInLocation("Monday, 02 January 2006 15:04", t1, serverLoc)
	}
	t, ok := p.document.Find(`meta[itemprop="dateCreated"]`).Attr("content")
	if ok {
		return time.ParseInLocation("2006-01-02 15:04:05", t, serverLoc)
	}
	return time.Time{}, fmt.Errorf("could not find publish date in %s", p.URL())
}

func (p *WhateleyPage) Tags() []StoryTag {
	sel := p.document.Find(`ul.tags li`)
	tags := make([]StoryTag, sel.Length())
	sel.Each(func(i int, s *goquery.Selection) {
		tags[i] = ParseTag(s)
	})
	return tags
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

// NOTE: Must end with a space, so that additional selectors can be concatenated on the end
const StoryBodySelector = `div.description.group div.desc-content.field_text `

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

var canonicalURLRegexp = regexp.MustCompile(`\Ahttp://whateleyacademy\.net/(?:index\.php/)?(?:content_page/)?([a-zA-Z0-9-]+)/(\d+)-([a-zA-Z0-9-]+)(?:\?|#|\z)`)
var idAndSlugRegexp = regexp.MustCompile(`(?:\A|/)(\d+)-([a-zA-Z%0-9-]+)(?:/|\z)`)

var stripExceptionsSelector = `
head base,
meta[name="rights"],
link[rel="canonical"],
meta[http-equiv="content-type"],
head title,
.flexi.element.field_published,
.flexi.element.field_created,
.flexi.element.field_modified,
.flexi.element.field_tags,
` + StoryBodySelector

func nodeInSelection(n *html.Node, s *goquery.Selection) bool {
	for {
		for _, v := range s.Nodes {
			if n == v {
				return true
			}
		}
		if n.Parent == nil {
			break
		}
		n = n.Parent
	}
	return false
}

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
	dontRemove = dontRemove.AddSelection(dontRemove.Parents())

	// CPU hog right here
	removed := page.document.Find("html *").NotSelection(dontRemove).RemoveMatcher(cascadia.Selector(func(n *html.Node) bool {
		if n.Type == html.TextNode {
			if nodeInSelection(n, dontRemove) {
				return false
			}
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

	canonicalLink, ok := doc.Find(`link[rel="canonical"]`).Attr("href")
	if !ok {
		return nil, errors.Errorf("could not find <link rel=canonical>")
	}
	urlShort, err := ParseURL(canonicalLink)
	if err != nil {
		return nil, err
	}
	page.StoryURL = urlShort

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
