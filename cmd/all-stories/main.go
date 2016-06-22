package main

import (
	"fmt"
	"strconv"
	"strings"

	"io/ioutil"
	"os"
	"sync"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"
	"gopkg.in/yaml.v2"
)

func fatal(err error) {
	fmt.Println("Fatal error:")
	fmt.Println(err.Error())
	os.Exit(2)
}

type result struct {
	Story     client.StoryURL
	WordCount int
}

func wordcountStory(ch chan<- result, storyID string, networkAccess *client.WANetwork) {
	fmt.Println(storyID)
	story, err := networkAccess.GetStoryByID(storyID)
	if err != nil {
		if strings.Contains(err.Error(), "Non-200 response") {
			return
		}
		if strings.Contains(err.Error(), "Could not parse canonical URL") {
			if strings.Contains(err.Error(), "/community") {
				return
			}
		}
		if strings.Contains(err.Error(), "Library stories are not supported") {
			return
		}
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(story.CategorySlug)
	if story.CategorySlug == "original-timeline" || story.CategorySlug == "stories" || story.CategorySlug == "2nd-gen-canon" {
		count := len(strings.Fields(story.StoryBodySelection().Text()))
		ch <- result{Story: story.StoryURL, WordCount: count}
	}
}

var skipIDs = []int{1, 4, 8, 9, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 26, 30, 358, 396, 641}

func emitAllIDs(idChan chan string, maxID int) {
	// Work Producer
outer:
	for i := 1; i < maxID; i++ {
		for _, v := range skipIDs {
			if i == v {
				continue outer
			}
		}
		idChan <- strconv.Itoa(i)
	}
	close(idChan)

}

func main() {

	networkAccess := client.New(client.Options{
		UserAgent: "Ebook tool - Count Words (+https://www.riking.org)",
		CacheFile: "./cache.db",
	})

	const maxID = 668
	const parallelLevel = 8
	// library: 434

	idChan := make(chan string)
	resChan := make(chan result)
	var wg sync.WaitGroup

	// Worker
	worker := func() {
		for v := range idChan {
			wordcountStory(resChan, v, networkAccess)
		}
		wg.Done()
	}

	wg.Add(parallelLevel)
	for i := 0; i < parallelLevel; i++ {
		go worker()
	}

	//go emitAllIDs(idChan, maxID)

	ebook := "gen2"

	var ebooksFile map[string]*ebooks.EpubDefinition
	b, err := ioutil.ReadFile(fmt.Sprintf("book-definitions/%s.yml", ebook))
	if err != nil {
		fatal(err)
	}
	err = yaml.Unmarshal(b, &ebooksFile)
	if err != nil {
		fatal(err)
	}
	go func() {
		b := ebooksFile[ebook]
		for _, v := range b.Parts {
			if v.Story.ID != "" {
				idChan <- v.Story.ID
			}
		}
		close(idChan)
	}()

	go func() {
		wg.Wait()
		close(resChan)
	}()

	total := 0
	// Consumer
	for v := range resChan {
		fmt.Printf("%d %s-%s\n", v.WordCount, v.Story.StoryID, v.Story.StorySlug)
		total += v.WordCount
	}

	fmt.Println("---------")
	fmt.Println("TOTAL:", total)
}
