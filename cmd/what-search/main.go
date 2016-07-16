package main

import (
	"flag"
	"fmt"

	"github.com/riking/whateley-ebooks/cmd"
)

func main() {
	networkAccess := cmd.Setup()
	networkAccess.UserAgent("Ebook tool - Search stories (+github.com/riking/whateley-ebooks)")

	search := flag.Arg(0)
	ids, err := networkAccess.SearchFulltext(search)
	for _, v := range ids {
		fmt.Println(v)
	}
	if err != nil {
		fmt.Println(err)
	}
}
