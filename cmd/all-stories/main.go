// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/cmd"
)

type result struct {
	client.StoryURL
	WordCount   int
	PublishDate time.Time
}

func wordcountStory(ch chan<- result, storyID string, networkAccess *client.WANetwork) {
	story, err := networkAccess.GetStoryByID(storyID)
	if err != nil {
		if strings.Contains(err.Error(), "fetching page HTML") {
			fmt.Fprintf(os.Stderr, "[W] Ignoring error fetching HTML for %s: %s\n", storyID, err)
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
	if story.CategorySlug == "original-timeline" || story.CategorySlug == "stories" || story.CategorySlug == "2nd-gen-canon" {
		count := len(strings.Fields(story.StoryBodySelection().Text()))
		date, err := story.PublishDate()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[F] Could not parse date for %s: %s\n", storyID, err)
			os.Exit(1)
		}
		ch <- result{StoryURL: story.StoryURL, WordCount: count, PublishDate: date}
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

func wordcountConsumer(resChan chan result) {
	total := 0
	for v := range resChan {
		fmt.Printf("%d %s-%s\n", v.WordCount, v.StoryID, v.StorySlug)
		total += v.WordCount
	}

	fmt.Println("---------")
	fmt.Println("TOTAL:", total)
}

type dateAndID struct {
	Published time.Time
	client.StoryURL
}

type sortByPubdate []dateAndID

func (a sortByPubdate) Len() int {
	return len(a)
}

func (a sortByPubdate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a sortByPubdate) Less(i, j int) bool {
	return a[i].Published.Before(a[j].Published)
}

func sortingConsumer(resChan chan result) {
	ary := make(sortByPubdate, 0, 150)

	for v := range resChan {
		ary = append(ary, dateAndID{Published: v.PublishDate, StoryURL: v.StoryURL})
		fmt.Fprintf(os.Stderr, ".")
	}

	sort.Sort(ary)

	fmt.Println("In publication order:")
	for _, v := range ary {
		fmt.Println(v.StoryID, v.StorySlug)
	}
}

func main() {
	// flag.String()

	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Examine All Stories (+github.com/riking/whateley-ebooks)")

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

	go emitAllIDs(idChan, maxID)

	//ebook := "gen2"
	//
	//var ebooksFile map[string]*ebooks.EpubDefinition
	//b, err := ioutil.ReadFile(fmt.Sprintf("book-definitions/%s.yml", ebook))
	//if err != nil {
	//	fatal(err)
	//}
	//err = yaml.Unmarshal(b, &ebooksFile)
	//if err != nil {
	//	fatal(err)
	//}
	//go func() {
	//	b := ebooksFile[ebook]
	//	for _, v := range b.Parts {
	//		if v.Story.ID != "" {
	//			idChan <- v.Story.ID
	//		}
	//	}
	//	close(idChan)
	//}()

	go func() {
		wg.Wait()
		close(resChan)
	}()

	//wordcountConsumer(resChan)
	sortingConsumer(resChan)
}
