package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/riking/whateley-ebooks/cmd"
	"github.com/riking/whateley-ebooks/ebooks"
)

func main() {
	_ = cmd.Setup()

	t := ebooks.GetAllTypos()
	b, err := json.Marshal(t)
	if err != nil {
		cmd.Fatal(err)
	}
	err = ioutil.WriteFile("typos.json", b, 0644)
	if err != nil {
		cmd.Fatal(err)
	}
}
