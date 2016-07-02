package cmd

import (
	"fmt"
	"os"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"
)

func Setup() *client.WANetwork {
	ebooks.SetTyposFromFile(ebooks.TyposDefaultFilename)
	networkAccess := client.New(client.Options{
		UserAgent: "(Error: tool name not specified) (+github.com/riking/whateley-ebooks)",
		CacheFile: "./cache.db",
	})

	return networkAccess
}

type causer interface {
	Cause() error
}

func Fatal(err error) {
	fmt.Println("Fatal error:")
	if _, ok := err.(causer); ok {
		fmt.Printf("%+v\n", err)
	} else {
		fmt.Println(err)
	}
	os.Exit(2)
}
