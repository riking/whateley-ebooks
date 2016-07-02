package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/riking/whateley-ebooks/cmd"
	"github.com/riking/whateley-ebooks/ebooks"
)

func main() {
	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Make EPub (+github.com/riking/whateley-ebooks)")

	flag.Parse()

	ebook := flag.Arg(0)
	if ebook == "" {
		fmt.Println("Please specify a book name / file on the command line")
		os.Exit(1)
	}

	var ebooksFile *ebooks.EpubDefinition
	attemptFiles := []string{ebook, fmt.Sprintf("book-definitions/%s.yml", ebook)}
	for _, v := range attemptFiles {
		_, err := os.Stat(v)
		if os.IsNotExist(err) {
			continue
		}

		b, err := ioutil.ReadFile(v)
		if err != nil {
			cmd.Fatal(errors.Wrapf(err, "Could not read %s", v))
		}
		err = yaml.Unmarshal(b, &ebooksFile)
		if err != nil {
			cmd.Fatal(errors.Wrapf(err, "Could not parse %s", v))
		}
	}

	var outFile string

	if flag.NArg() == 2 {
		outFile = flag.Arg(1)
		// TODO(riking): allow stdout output? need to make CreateEpub take io.Writer
	} else {
		err := os.Mkdir("target", 0755)
		if !os.IsExist(err) {
			cmd.Fatal(errors.Wrap(err, "Could not create output directory"))
		}
		outFile = fmt.Sprintf("target/%s.epub", ebook)
	}

	err := ebooksFile.Prepare(networkAccess)
	if err != nil {
		cmd.Fatal(err)
	}

	err = ebooks.CreateEpub(ebooksFile, networkAccess, outFile)
	if err != nil {
		cmd.Fatal(err)
	}
}
