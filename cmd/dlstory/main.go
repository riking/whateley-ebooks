// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main // import "github.com/riking/whateley-ebooks/cmd/dlstory"

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/cmd"
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

var purgeCache = flag.Bool("purge", false, "Purge cache for the page")

func main() {
	// flag.String()

	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - TyposFile testing (+github.com/riking/whateley-ebooks)")

	storyID := flag.Arg(0)
	if storyID == "" {
		fmt.Println("Please specify a story ID on the command line")
		os.Exit(1)
	}

	if *purgeCache == true {
		acted, err := networkAccess.PurgeCache(client.StoryURL{StoryID: storyID})
		if err != nil {
			cmd.Fatal(err)
		}
		if acted {
			fmt.Println("Cache purged..")
		}
	}

	page, err := networkAccess.GetStoryByID(storyID)
	if err != nil {
		cmd.Fatal(err)
	}
	fmt.Println(page.URL())
	fmt.Println(page.PublishDate())
	tags := page.Tags()
	for _, v := range tags {
		fmt.Printf("%s %s - %s\n", v.ID, v.Slug, v.Name)
	}
	ioutil.WriteFile("dlstory-before.html", []byte(page.StoryBody()), 0644)
	ebooks.FixForEbook(page)
	ioutil.WriteFile("dlstory-after.html", []byte(page.StoryBody()), 0644)
	html, _ := page.Doc().Html()
	ioutil.WriteFile("dlstory-full.html", []byte(html), 0644)
}
