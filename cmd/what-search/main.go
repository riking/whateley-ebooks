package main

import (
	"flag"
	"fmt"

	"github.com/PuerkitoBio/goquery"
	"github.com/riking/whateley-ebooks/cmd"
	"os"
	"regexp"
	"sort"
	"strconv"
)

func main() {
	charsAfter := flag.Int("A", 15, "characters after the match to include")
	charsBefore := flag.Int("B", 15, "characters before the match to include")
	charsAround := flag.Int("C", -1, "characters around the match to include (combination of -A and -B)")

	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Search stories (+github.com/riking/whateley-ebooks)")

	if *charsAround != -1 {
		*charsAfter = *charsAround
		*charsBefore = *charsAround
	}

	search := flag.Arg(0)
	ids, err := networkAccess.SearchFulltext(search)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	sort.Ints(ids)

	searchRgx := regexp.MustCompile(regexp.QuoteMeta(search))
	for _, id := range ids {
		st, err := networkAccess.GetStoryByID(strconv.Itoa(id))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR] %3d: %s\n", id, err)
			continue
		}
		html, err := goquery.OuterHtml(st.Doc().Find("html"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERR] %3d: %s\n", id, err)
			continue
		}
		matches := searchRgx.FindAllStringIndex(html, -1)

		for _, m := range matches {
			startIdx := m[0] - *charsBefore
			endIdx := m[1] + *charsAfter
			matchStr := html[startIdx:endIdx]
			fmt.Printf("[ M ] %3d: %s\n", id, matchStr)
		}
	}
}
