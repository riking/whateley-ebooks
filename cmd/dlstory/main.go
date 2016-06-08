// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main // import "github.com/riking/whateley-ebooks/cmd/dlstory"

import (
	"fmt"
	"io/ioutil"
	"os"

	"flag"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"
)

func getPage(url string, access *client.WANetwork) (*client.WhateleyPage, error) {
	u, _ := client.ParseURL(url)
	return access.GetStoryByID(u.StoryID)
}

func fatal(err error) {
	fmt.Println("Fatal error:")
	fmt.Println(err.Error())
	os.Exit(2)
}

func main() {
	ebooks.SetTyposFromFile(ebooks.TyposDefaultFilename)

	networkAccess := client.New(client.Options{
		UserAgent: "Ebook tool - TyposFile testing (+https://www.riking.org)",
		CacheFile: "./cache.db",
	})

	flag.Parse()

	storyID := flag.Arg(0)
	if storyID == "" {
		fmt.Println("Please specify a story ID on the command line")
		os.Exit(1)
	}

	page, err := networkAccess.GetStoryByID(storyID)
	if err != nil {
		fatal(err)
	}
	fmt.Println(page.URL())
	fmt.Println(page.PublishDate())
	ioutil.WriteFile("out1.html", []byte(page.StoryBody()), 0644)
	ebooks.FixForEbook(page)
	ioutil.WriteFile("out2.html", []byte(page.StoryBody()), 0644)
	html, _ := page.Doc().Html()
	ioutil.WriteFile("out0", []byte(html), 0644)
}
