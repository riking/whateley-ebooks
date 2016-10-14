// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package ebooks // import "github.com/riking/whateley-ebooks/ebooks"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/cascadia"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v2"

	"github.com/riking/whateley-ebooks/client"
)

// TypoFix represents a single fixup processing for a story.
type TypoFix struct {
	FindSelector    string   `yaml:"select,omitempty" json:"select,omitempty"`
	FindText        string   `yaml:",omitempty" json:"findtext,omitempty"`
	FindHTML        string   `yaml:",omitempty" json:"findhtml,omitempty"`
	FindModifiers   []string `yaml:"selectmod,omitempty" json:"selectmod,omitempty"`
	ReplaceSelector string   `yaml:",omitempty" json:"replaceselector,omitempty"`
	ReplaceText     string   `yaml:",omitempty" json:"replacetext,omitempty"`
	ReplaceHTML     string   `yaml:"replace,omitempty" json:"replace,omitempty"`
	X_ReplaceHTML2  string   `yaml:"replacehtml,omitempty" json:"replacehtml,omitempty"`
	Attribute       string   `yaml:"attr,omitempty" json:"attr,omitempty"`
	Action          string   `yaml:",omitempty" json:"action,omitempty"`
	Include         string   `yaml:"include,omitempty" json:"include,omitempty"`
	OnlyWhen        string   `yaml:"onlywhen,omitempty" json:"onlywhen,omitempty"`
}

func (t TypoFix) Find(doc *goquery.Document) *goquery.Selection {
	var s *goquery.Selection
	if t.FindHTML != "" {
		panic("findHTML not implemented")
	}
	if t.FindSelector != "" {
		s = doc.Find(client.StoryBodySelector + t.FindSelector)
	} else if t.FindText != "" {
		s = doc.Find(client.StoryBodySelector + " p:contains(\"" + t.FindSelector + "\")")
	}

	for _, v := range t.FindModifiers {
		switch v {
		case "parent":
			s = s.Parent()
		case "addNextSibling":
			s = s.AddSelection(s.Next())
		case "nextSibling":
			s = s.Next()
		}
	}
	return s
}

func (t TypoFix) Apply(p *client.WhateleyPage) {
	if t.OnlyWhen != "" {
		spl := strings.Split(t.OnlyWhen, ",")
		skip := true
		for _, v := range spl {
			if v == "ebook" {
				skip = false
				break
			}
		}

		if skip {
			return
		}
	}
	if t.FindText != "" && t.ReplaceText != "" {
		t.Action = "replacetext"
	}
	if t.X_ReplaceHTML2 != "" {
		t.ReplaceHTML = t.X_ReplaceHTML2
	}

	switch t.Action {
	case "unwrap":
		t.Find(p.Doc()).Unwrap()
	case "wrap":
		t.Find(p.Doc()).WrapHtml(t.ReplaceHTML)
	case "wrapAll":
		t.Find(p.Doc()).WrapAllHtml(t.ReplaceHTML)
	case "wrapInner":
		t.Find(p.Doc()).WrapInnerHtml(t.ReplaceHTML)
	case "deleteAttr":
		t.Find(p.Doc()).RemoveAttr(t.Attribute)
	case "setAttr":
		t.Find(p.Doc()).SetAttr(t.Attribute, t.ReplaceHTML)
	case "insertBefore":
		t.Find(p.Doc()).BeforeHtml(t.ReplaceHTML)
	case "replacetext":
		t.Find(p.Doc()).Each(func(_ int, s *goquery.Selection) {
			html, err := goquery.OuterHtml(s)
			if err != nil {
				panic(err)
			}
			s.ReplaceWithHtml(strings.Replace(html, t.FindText, t.ReplaceText, -1))
		})
	case "replacehtml":
		s := t.Find(p.Doc())
		s.ReplaceWithHtml(t.ReplaceHTML)
	case "paragraphsToLinebreaks":
		sel := t.Find(p.Doc())
		sel.Each(func(_ int, s *goquery.Selection) {
			para := s.Get(0)
			newNode, err := html.ParseFragment(bytes.NewBufferString("<br>"), para)
			if err != nil {
				panic(err)
			}
			para.AppendChild(newNode[0])
		})
		sel.Find("p > *").Unwrap()
	default:
		fmt.Fprintf(os.Stderr, "[ebooks] warning: unknown typos.yml action %s\n", t.Action)
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

func GetAllTypos() *TyposFile {
	return &typosFile
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
		fmt.Printf("#%s: %d typos\n", p.StoryID, 0)
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

	fmt.Printf("#%s: %d typos\n", p.StoryID, len(typos))
	return typos
}

//

var hrSelectors = []string{
	`* > img[src="/images/breaks/linebreak-bluearcs.jpg"]`,
	`* > img[src="/images/hr1.gif"]`,
	// `center > img[src="/images/breaks/linebreak-bluearcs.jpg"]`,
	`center > img[alt="linebreak shadow"]`,
	`div.hr`,
	`div.hr2`,
	`hr[style]`,
}
var hrParagraphs = []string{
	" ",
	"\u00a0",
	"*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*",
}

var hrParagraphRegex *regexp.Regexp
var hrParagraphRegexLock sync.Mutex

func getHrParagraphRegex() *regexp.Regexp {
	hrParagraphRegexLock.Lock()
	defer hrParagraphRegexLock.Unlock()
	if hrParagraphRegex != nil {
		return hrParagraphRegex
	}
	var buf bytes.Buffer
	buf.WriteString("\\A(")
	buf.WriteString(`[\*\p{Zs}]+`)
	buf.WriteRune('|')
	for i, v := range hrParagraphs {
		buf.WriteString(regexp.QuoteMeta(v))
		if i != len(hrParagraphs)-1 {
			buf.WriteRune('|')
		}
	}
	buf.WriteString(")\\z")
	hrParagraphRegex = regexp.MustCompile(buf.String())

	quickTest := []bool{
		hrParagraphRegex.MatchString("* * * *"),
		hrParagraphRegex.MatchString(" * * * * "),
		hrParagraphRegex.MatchString(" "),
		hrParagraphRegex.MatchString("\u00a0"),
		hrParagraphRegex.MatchString("*\u00a0*\u00a0*\u00a0*"),
		hrParagraphRegex.MatchString("*\u00a0 *\u00a0 *\u00a0 *"),
		hrParagraphRegex.MatchString("*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0 *\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0\u00a0*"),
		hrParagraphRegex.MatchString("\u00a0*\u00a0*\u00a0*\u00a0*\u00a0"),
	}

	for i, v := range quickTest {
		if !v {
			panic(fmt.Sprintf("<hr> failed self-test #%d", i))
		}
	}

	return hrParagraphRegex
}

func hrParagraphMatcher() func(*html.Node) bool {
	paraRegex := getHrParagraphRegex()
	return func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		if n.Data != "p" && n.Data != "div" && n.Data != "strong" && n.Data != "span" {
			return false
		}
		d := goquery.NewDocumentFromNode(n)
		return paraRegex.MatchString(d.Text())
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
	for _, v := range getTypos(p) {
		v.Apply(p)
	}
}

func FixForEbook(p *client.WhateleyPage) error {
	var s *goquery.Selection

	// Fix \u0012 and friends
	html, _ := goquery.OuterHtml(p.StoryBodySelection())
	p.StoryBodySelection().ReplaceWithHtml(
		strings.Replace(strings.Replace(strings.Replace(
			strings.Replace(html, "\u0012", "’", -1),
			"\u0016", "—", -1),
			"\u0005", "…", -1),
			"oe\u001C", "œ", -1))

	// Apply typo corrections
	applyTypos(p)

	// Fix horizontal rules
	s = p.Doc().Find("")
	for _, sel := range hrSelectors {
		s = s.Add(client.StoryBodySelector + sel)
	}
	s = s.AddMatcher(cascadia.Selector(hrParagraphMatcher()))
	s.ReplaceWithHtml("<hr>")

	p.Doc().Find("p hr").Parent().ReplaceWithHtml("<hr>")
	p.Doc().Find("center hr").Parent().ReplaceWithHtml("<hr>")

	// Fix double hrs
	p.Doc().Find("hr + hr").Remove()

	//p.Doc().Find("blockquote .lyrics .PCscreen").Unwrap()

	return nil
}
