// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main // import "github.com/riking/whateley-ebooks/cmd/dlstory"

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"

	"github.com/pkg/errors"
)

var httpClient *client.Client

func getPage(url string) (*client.WhateleyPage, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("constructing request to %s", url))
	}
	doc, err := httpClient.Document(req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("fetching document at %s", url))
	}
	return client.ParseStoryPage(doc)
}

func fatal(err error) {
	fmt.Println("Fatal error:")
	fmt.Println(err.Error())
	os.Exit(2)
}

func main() {
	ebooks.SetTyposFromFile(ebooks.TyposDefaultFilename)

	httpClient = client.New(client.Options{
		UserAgent: "ebooks tool test script (+https://www.riking.org)",
		CacheFile: "./cache.db",
	})

	url := `http://whateleyacademy.net/index.php/wrong-category/208-wrong-slug`
	url = `http://whateleyacademy.net/index.php/stories/279-hive-part-4-who-dun-it`
	page, err := getPage(url)
	if err != nil {
		fatal(err)
	}
	fmt.Println(page.URL())
	ioutil.WriteFile("out1", []byte(page.StoryBody()), 0644)
	ebooks.FixForEbook(page)
	ioutil.WriteFile("out2", []byte(page.StoryBody()), 0644)
	html, _ := page.Doc().Html()
	ioutil.WriteFile("out0", []byte(html), 0644)
}
