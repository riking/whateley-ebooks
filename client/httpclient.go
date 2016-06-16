// Copyright © Kane York 2016.
// Please see COPYRIGHT.md and LICENSE-CODE.txt.

package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/peterbourgon/diskv"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	Headers    http.Header
	httpClient http.Client
	options    Options
	cache      *diskv.Diskv
}

type Options struct {
	CacheDir  string
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

func getCacheKey(url string) string {
	fields, err := ParseURL(url)
	if err != nil {
		return url
	}
	fields.StorySlug = "x"
	return fields.CacheKey()
}

func getCacheKeyURL(u *url.URL) string {
	return getCacheKey(u.String())
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

	if opts.CacheDir != "" {
		c.cache = diskv.New(diskv.Options{
			BasePath:     opts.CacheDir,
			CacheSizeMax: 1024 * 1024 * 300,
		})
	}
	return c
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	for k := range c.Headers {
		req.Header.Set(k, c.Headers.Get(k))
	}
	return c.httpClient.Do(req)
}

func (c *Client) Document(req *http.Request) (*goquery.Document, error) {
	cacheKey := getCacheKeyURL(req.URL)
	if c.cache.Has(cacheKey) {
		fmt.Println("cache hit:", cacheKey)
		r, err := c.cache.ReadStream(cacheKey, false)
		if err != nil {
			return nil, err
		}
		return goquery.NewDocumentFromReader(r)
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

	err = c.cache.Write(cacheKey, b)
	buf := bytes.NewBuffer(b)

	return goquery.NewDocumentFromReader(buf)
}
