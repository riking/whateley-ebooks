package main

import (
	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"

	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func fatal(err error) {
	fmt.Println("Fatal error:")
	fmt.Println(err.Error())
	os.Exit(2)
}

func main() {
	ebooks.SetTyposFromFile(ebooks.TyposDefaultFilename)

	networkAccess := client.New(client.Options{
		UserAgent: "Ebook tool - Make EPub (+https://www.riking.org)",
		CacheFile: "./cache.db",
	})

	ebook := "personal-nextfew"

	var ebooksFile map[string]*ebooks.EpubDefinition
	b, err := ioutil.ReadFile(fmt.Sprintf("book-definitions/%s.yml", ebook))
	if err != nil {
		fatal(err)
	}
	err = yaml.Unmarshal(b, &ebooksFile)
	if err != nil {
		fatal(err)
	}

	os.Mkdir("target", 0755)
	filename := fmt.Sprintf("target/%s.epub", ebook)
	err = ebooks.CreateEpub(ebooksFile[ebook], networkAccess, filename)
	if err != nil {
		fatal(err)
	}
}
