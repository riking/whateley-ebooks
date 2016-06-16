// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package ebooks // import "github.com/riking/whateley-ebooks/ebooks"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/riking/whateley-ebooks/client"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"
)

// TypoFix represents a single fixup processing for a story.
type TypoFix struct {
	FindSelector    string `yaml:"select,omitempty"`
	FindText        string `yaml:",omitempty"`
	FindHTML        string `yaml:",omitempty"`
	ReplaceSelector string `yaml:",omitempty"`
	ReplaceText     string `yaml:",omitempty"`
	ReplaceHTML     string `yaml:"replace,omitempty"`
	Attribute       string `yaml:"attr,omitempty"`
	Action          string `yaml:",omitempty"`
	Include         string `yaml:"include,omitempty"`
}

func (t TypoFix) Find(doc *goquery.Document) *goquery.Selection {
	var s *goquery.Selection
	if t.FindHTML != "" {
		panic("findHTML not implemented")
	}
	if t.FindSelector != "" {
		if t.FindText != "" {
			s = doc.Find(fmt.Sprintf("%s %s:contains(\"%s\")", client.StoryBodySelector, t.FindSelector, t.FindText))
		} else {
			s = doc.Find(client.StoryBodySelector + t.FindSelector)
		}
	} else if t.FindText != "" {
		s = doc.Find(client.StoryBodySelector + " p:contains(\"" + t.FindSelector + "\")")
	}
	return s
}

func (t TypoFix) Apply(p *client.WhateleyPage) {
	switch t.Action {
	case "unwrap":
		t.Find(p.Doc()).Unwrap()
	case "wrap":
		t.Find(p.Doc()).WrapHtml(t.ReplaceHTML)
	case "deleteAttr":
		t.Find(p.Doc()).RemoveAttr(t.Attribute)
	default:
		fmt.Printf("[ebooks] warning: unknown typos.yml action %s\n", t.Action)
	}
}

// TyposFile is the file format for typos.yml.
// The key of the map is the story ID string.
type TyposFile struct {
	ByStoryID map[string][]TypoFix `yaml:",inline"`
	Library   map[string][]TypoFix `yaml:"library"`
}

const TyposDefaultFilename = "./typos.yml"

var typosFile TyposFile

func SetTypos(t TyposFile) {
	typosFile = t
}

func SetTyposFromFile(filename string) error {
	if filename == "" {
		filename = TyposDefaultFilename
	}
	t := TyposFile{
		ByStoryID: make(map[string][]TypoFix),
		Library:   make(map[string][]TypoFix),
	}
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "could not read typos file")
	}
	err = yaml.Unmarshal(bytes, &t)
	if err != nil {
		return errors.Wrap(err, "syntax error in typos file")
	}
	typosFile = t
	return nil
}

func getTypos(p *client.WhateleyPage) []TypoFix {
	if p.StorySlug == "" {
		panic("story slug was empty string")
	}

	oTypos, ok := typosFile.ByStoryID[p.StoryID]
	if !ok {
		return nil
	}

	// Process include: statements
	typos := make([]TypoFix, 0, len(oTypos))
	for _, v := range oTypos {
		if v.Include != "" {
			for _, v2 := range typosFile.Library[v.Include] {
				typos = append(typos, v2)
			}
		} else {
			typos = append(typos, v)
		}
	}
	return typos
}

//

var hrSelectors = []string{
	`p > img[alt="linebreak bluearcs"]`,
	`div.hr`,
	`div.hr2`,
	`hr[style]`,
}
var hrParagraphs = []string{
	"\u00a0",
	"*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0 *\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*",
	"*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*",
}

var hrParagraphRegex *regexp.Regexp

var h3Selectors = []string{
	`p.lyrics strong em`,
	`p.lyrics em strong`,
}

func getHrParagraphRegex() *regexp.Regexp {
	if hrParagraphRegex != nil {
		return hrParagraphRegex
	}
	var buf bytes.Buffer
	buf.WriteString("\\A(")
	// " " is a nbsp
	buf.WriteString(`\*((\s)+\*)+`)
	buf.WriteRune('|')
	for i, v := range hrParagraphs {
		buf.WriteString(regexp.QuoteMeta(v))
		if i != len(hrParagraphs)-1 {
			buf.WriteRune('|')
		}
	}
	buf.WriteString(")\\z")
	fmt.Println(buf.String())
	hrParagraphRegex = regexp.MustCompile(buf.String())

	return hrParagraphRegex
}

func hrParagraphMatcher() func(*html.Node) bool {
	paraRegex := getHrParagraphRegex()
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		if n.Data != "p" {
			return false
		}
		d := goquery.NewDocumentFromNode(n)
		html, err := d.Html()
		if err != nil {
			panic(errors.Wrap(err, "error returned from Html()"))
		}
		return paraRegex.MatchString(html)
	}
}

func searchRegexp(search *regexp.Regexp) func(*html.Node) bool {
	return func(n *html.Node) bool {
		if n.Type != html.TextNode {
			return false
		}
		return search.MatchString(strings.TrimSpace(n.Data))
	}
}

func applyTypos(p *client.WhateleyPage) {
	//curHtml, err := goquery.OuterHtml(p.StoryBodySelection())
	//if err != nil {
	//	panic(errors.Wrap(err, "could not convert storybody to html"))
	//}
	for _, v := range getTypos(p) {
		v.Apply(p)
	}
}

func FixForEbook(p *client.WhateleyPage) error {
	var s *goquery.Selection

	// Apply typo corrections
	applyTypos(p)

	// Fix horizontal rules
	s = p.Doc().Find("")
	for _, sel := range hrSelectors {
		s = s.Add(client.StoryBodySelector + sel)
	}
	s = s.AddMatcher(cascadia.Selector(hrParagraphMatcher()))
	fmt.Println(s.Length())
	hrsReplaced := s.ReplaceWithHtml("<hr>")
	fmt.Println("replaced", hrsReplaced.Length(), "<hr>s")
	hrsReplaced.Each(func(_ int, s *goquery.Selection) {
		fmt.Println(s.Html())
	})

	s = p.Doc().Find("")
	for _, sel := range h3Selectors {
		s.Add(client.StoryBodySelector + sel)
	}
	s.WrapAll("<h3>")

	return nil
}
