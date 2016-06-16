// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package client

import (
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	Headers    http.Header
	httpClient http.Client
	options    Options
	db         *sql.DB
}

type Options struct {
	CacheFile string
	UserAgent string
	Headers   http.Header
}

type printingRoundTripper struct {
	parent http.RoundTripper
}

func (p *printingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if p.parent == nil {
		p.parent = http.DefaultTransport
	}

	fmt.Printf("> %s %s\n", req.Method, req.URL.String())
	resp, err := p.parent.RoundTrip(req)
	fmt.Printf("< %s %s\n", req.Method, req.URL.String())
	return resp, err
}

func New(opts Options) *Client {
	c := new(Client)
	if opts.UserAgent == "" {
		opts.UserAgent = "Client Name Not Set (+github.com/riking/whateley)"
	}
	c.Headers = opts.Headers
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	c.Headers.Set("User-Agent", opts.UserAgent)
	c.httpClient.Jar, _ = cookiejar.New(nil)
	c.httpClient.Timeout = 15 * time.Second
	c.httpClient.Transport = &printingRoundTripper{c.httpClient.Transport}

	if opts.CacheFile != "" {
		conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s", opts.CacheFile))
		if err != nil {
			panic(err)
		}
		c.db = conn
		err = c.setupDB()
		if err != nil {
			panic(err)
		}
	}
	return c
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for k := range c.Headers {
		req.Header.Set(k, c.Headers.Get(k))
	}
	return c.httpClient.Do(req)
}

// Document gets a URL, cached, and returns a goquery.Document.
func (c *Client) Document(req *http.Request) (*goquery.Document, error) {
	if req.Method != "GET" {
		panic("Document() does not support non-GET")
	}

	u, err := ParseURL(req.URL.String())
	switch err {
	case nil:
		id, err := c.cacheCheck(u)
		if err != nil {
			return nil, errors.Wrap(err, "checking cache for page")
		}
		if id == -1 {
			// cache miss
			break
		}

		b, err := c.cacheGet(id)
		if err != nil {
			return nil, errors.Wrap(err, "Retrieving value from cache")
		}
		return goquery.NewDocumentFromReader(bytes.NewBuffer(b))
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// if time is OK
	err = c.cachePut(u, b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[db] warning: could not add to cache: %s", err)
	}

	buf := bytes.NewBuffer(b)

	return goquery.NewDocumentFromReader(buf)
}
