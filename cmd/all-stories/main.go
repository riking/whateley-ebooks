// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/cmd"
)

const maxID = 800

type result struct {
	client.StoryURL
	WordCount   int
	PublishDate time.Time
}

func getStory(ch chan<- *client.WhateleyPage, storyID string, networkAccess *client.WANetwork) {
	story, err := networkAccess.GetStoryByID(storyID)
	if err != nil {
		if strings.Contains(err.Error(), "fetching page HTML") {
			if strings.Contains(err.Error(), "404 for http") {
				fmt.Fprint(os.Stderr, "4")
				return
			}
			if strings.Contains(err.Error(), "403 for http") {
				fmt.Fprint(os.Stderr, "3")
				return
			}
			fmt.Fprintf(os.Stderr, "\n[W] Ignoring error fetching HTML for %s: %s\n", storyID, err)
			return
		}
		if strings.Contains(err.Error(), "Could not parse canonical URL") {
			if strings.Contains(err.Error(), "/community") {
				fmt.Fprint(os.Stderr, "C")
				return
			}
		}
		fmt.Println(err)
		os.Exit(1)
	}

	categoryAccepted := *includeEverything
	if *includeGen1 && (story.CategorySlug == "original-timeline" || story.CategorySlug == "stories") {
		categoryAccepted = true
	}
	if *includeGen2 && story.CategorySlug == "2nd-gen-canon" {
		categoryAccepted = true
	}
	if *includeFanFiction && story.CategorySlug == "featured-fan-fiction" {
		categoryAccepted = true
	}
	if *includeLibrary && (story.CategorySlug == "the-library" || strings.HasPrefix(story.CategorySlug, "the-library/")) {
		categoryAccepted = true
	}

	if !categoryAccepted {
		fmt.Fprintf(os.Stderr, "\n[W] Ignoring page %s with category %s\n", story.StoryID, story.CategorySlug)
		return
	}
	ch <- story
}

func noopProcess(ch chan<- result, story *client.WhateleyPage, networkAccess *client.WANetwork) {
	//date, err := story.PublishDate()
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "\n[F] Could not parse date for %s: %s\n", story.StoryID, err)
	//	os.Exit(1)
	//}
	ch <- result{StoryURL: story.StoryURL}
}

func wordcountProcess(ch chan<- result, story *client.WhateleyPage, networkAccess *client.WANetwork) {
	count := len(strings.Fields(story.StoryBodySelection().Text()))
	date, err := story.PublishDate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[F] Could not parse date for %s: %s\n", story.StoryID, err)
		os.Exit(1)
	}
	ch <- result{StoryURL: story.StoryURL, WordCount: count, PublishDate: date}
}

type categoryPair struct {
	Text    string
	FromURL string
	Href    string
}

var allCategories = make(map[categoryPair]struct{})
var allTags = make(map[client.StoryTag]struct{})
var allAuthors = make(map[string]struct{})

func recordUniqueProcess(ch chan<- result, story *client.WhateleyPage, networkAccess *client.WANetwork, extra *map[string]interface{}) {
	catMap := (*extra)["c"].(map[categoryPair]struct{})
	tagMap := (*extra)["t"].(map[client.StoryTag]struct{})
	authorMap := (*extra)["a"].(map[string]struct{})
	cat := categoryPair{
		FromURL: story.CategorySlug,
		Text:    story.Category(),
		Href:    story.CategoryLink(),
	}
	if cat.FromURL == "-" || cat.Text == "" || cat.Href == "" {
		fmt.Println("############## EMPTY CATEGORY", story.StoryURL.StoryID, story.StoryURL.URL())
	}
	catMap[cat] = struct{}{}
	for _, v := range story.Tags() {
		tagMap[v] = struct{}{}
	}
	authorMap[story.Authors()] = struct{}{}
	noopProcess(ch, story, networkAccess)
}

//var skipIDs = []int{1, 4, 8, 9, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 26, 30, 358, 396, 641, 672}
var skipIDs = []int{
	672, // Is /published, its category is /15-public-news
	680, // Is /chat, its category is /36-empty
}

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

type wcAndID struct {
	wordcount int
	client.StoryURL
}
type sortByWordcount []wcAndID

func (a sortByWordcount) Len() int {
	return len(a)
}

func (a sortByWordcount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a sortByWordcount) Less(i, j int) bool {
	return a[i].wordcount < a[j].wordcount
}
func wordcountConsumer(resChan chan result) {
	ary := make(sortByWordcount, 0)
	total := 0
	for v := range resChan {
		fmt.Fprintf(os.Stderr, "%d %s-%s\n", v.WordCount, v.StoryID, v.StorySlug)
		ary = append(ary, wcAndID{wordcount: v.WordCount, StoryURL: v.StoryURL})
		total += v.WordCount
	}

	sort.Sort(ary)

	for _, v := range ary {
		fmt.Fprintf(os.Stdout, "%d %s-%s\n", v.wordcount, v.StoryID, v.StorySlug)
	}

	fmt.Fprintln(os.Stderr, "---------")
	fmt.Fprintln(os.Stderr, "TOTAL:", total)
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
		fmt.Fprint(os.Stderr, ".")
	}

	sort.Sort(ary)

	fmt.Println("In publication order:")
	for _, v := range ary {
		fmt.Println(v.URL())
	}
}

func collectingConsumer(resChan chan result) []result {
	ary := make([]result, 0, 150)

	for v := range resChan {
		ary = append(ary, v)
		fmt.Fprint(os.Stderr, ".")
	}

	return ary
}

var includeEverything = flag.Bool("everything", false, "Process everything")
var includeLibrary = flag.Bool("library", false, "Include library stories")
var includeFanFiction = flag.Bool("fanfiction", false, "Include fanfiction")
var includeGen2 = flag.Bool("gen2", true, "Include gen2")
var includeGen1 = flag.Bool("gen1", true, "Include gen1")

var cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	// flag.String()

	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Examine All Stories (+github.com/riking/whateley-ebooks)")

	fmt.Fprintf(os.Stderr, "Gen1: %s", *includeGen1)
	fmt.Fprintf(os.Stderr, "Gen2: %s", *includeGen2)
	fmt.Fprintf(os.Stderr, "Library: %s", *includeLibrary)
	fmt.Fprintf(os.Stderr, "Fan Fiction: %s", *includeFanFiction)
	//networkAccess.DBTest()
	//return

	workFunc := recordUniqueProcess

	const parallelLevel = 8
	// library: 434

	idChan := make(chan string)
	storyChan := make(chan *client.WhateleyPage)
	resChan := make(chan result)
	var fetchWg sync.WaitGroup
	var procWg sync.WaitGroup

	fetchWorker := func() {
		for v := range idChan {
			getStory(storyChan, v, networkAccess)
		}
		fetchWg.Done()
	}

	fetchWg.Add(parallelLevel)
	for i := 0; i < parallelLevel; i++ {
		go fetchWorker()
	}

	extraDatas := make([]map[string]interface{}, parallelLevel)
	processWorker := func(i int) {
		extraData := make(map[string]interface{})
		extraData["c"] = make(map[categoryPair]struct{})
		extraData["t"] = make(map[client.StoryTag]struct{})
		extraData["a"] = make(map[string]struct{})
		for v := range storyChan {
			workFunc(resChan, v, networkAccess, &extraData)
		}
		extraDatas[i] = extraData
		procWg.Done()
	}

	procWg.Add(parallelLevel)
	for i := 0; i < parallelLevel; i++ {
		go processWorker(i)
	}

	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			cmd.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	go emitAllIDs(idChan, maxID)

	go func() {
		fetchWg.Wait()
		close(storyChan)
	}()
	go func() {
		procWg.Wait()
		close(resChan)
	}()

	//wordcountConsumer(resChan)
	sortingConsumer(resChan)
	return
	allStoryUrls := collectingConsumer(resChan)

	// combine per-goroutine maps
	for _, ed := range extraDatas {
		fmt.Println(ed)
		for k, v := range ed["c"].(map[categoryPair]struct{}) {
			allCategories[k] = v
		}
		for k, v := range ed["t"].(map[client.StoryTag]struct{}) {
			allTags[k] = v
		}
		for k, v := range ed["a"].(map[string]struct{}) {
			allAuthors[k] = v
		}
	}

	const separator = "============================================================"
	fmt.Println(separator)

	for _, v := range allStoryUrls {
		fmt.Println(v.URL())
	}
	fmt.Println(separator)

	for k := range allAuthors {
		fmt.Println(k)
	}
	fmt.Println(separator)

	for k := range allCategories {
		fmt.Println(k)
	}
	fmt.Println(separator)

	for k := range allTags {
		fmt.Println(k)
	}
}
