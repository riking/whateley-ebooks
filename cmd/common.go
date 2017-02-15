// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/riking/whateley-ebooks/client"
	"github.com/riking/whateley-ebooks/ebooks"
)

var offlineMode *bool

func Setup() *client.WANetwork {
	ebooks.SetTyposFromFile("./typos.yml")
	offlineMode = flag.Bool("offline", false, "Operate in offline mode (cached entries never expire).")
	maxRequests := flag.Int("max-requests", 10, "Maximum number of concurrent outstanding HTTP requests")

	flag.Parse()

	networkAccess := client.New(client.Options{
		UserAgent:      "(Error: tool name not specified) (+github.com/riking/whateley-ebooks)",
		CacheFile:      "./cache.db",
		Offline:        *offlineMode,
		MaxConcurrency: *maxRequests,
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
