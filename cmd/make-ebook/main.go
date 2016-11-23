// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/cmd"
	"github.com/riking/whateley-ebooks/ebooks"
)

func createEbook(bookID string, networkAccess *client.WANetwork) error {
	var ebooksFile *ebooks.EpubDefinition
	attemptFiles := []string{bookID, fmt.Sprintf("book-definitions/%s", bookID), fmt.Sprintf("book-definitions/%s.yml", bookID)}
	for _, v := range attemptFiles {
		_, err := os.Stat(v)
		if os.IsNotExist(err) {
			continue
		}

		b, err := ioutil.ReadFile(v)
		if err != nil {
			return errors.Wrapf(err, "Could not read %s", v)
		}
		err = yaml.Unmarshal(b, &ebooksFile)
		if err != nil {
			return errors.Wrapf(err, "Could not parse %s", v)
		}
	}

	var outFile string = fmt.Sprintf("target/%s.epub", strings.TrimSuffix(path.Base(bookID), ".yml"))

	err := ebooksFile.Prepare(networkAccess)
	if err != nil {
		return errors.Wrapf(err, "Failed to prepare %s", bookID)
	}

	err = ebooks.CreateEpub(ebooksFile, networkAccess, outFile)
	if err != nil {
		return errors.Wrapf(err, "Failed to create %s", bookID)
	}
	return nil
}

func main() {
	// flag.String()

	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Make EPub (+github.com/riking/whateley-ebooks)")

	bookIDs := flag.Args()
	if len(bookIDs) == 0 {
		fmt.Println("Please specify a book name / file on the command line")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	var errs []error
	wg.Add(len(bookIDs))
	errs = make([]error, len(bookIDs))

	fmt.Println(bookIDs)

	for i := range bookIDs {
		go func(idx int) {
			bookID := bookIDs[idx]
			err := createEbook(bookID, networkAccess)
			errs[idx] = err
			wg.Done()
		}(i)
	}
	wg.Wait()

	for i, v := range bookIDs {
		if errs[i] == nil {
			fmt.Printf("[ OK] %s\n", v)
		} else {
			fmt.Printf("[ERR] %s: %s\n", v, errs[i])
		}
	}
}
