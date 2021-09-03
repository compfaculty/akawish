package main

import (
	"github.com/mmcdole/gofeed"
	"math/rand"
	"time"
)

type Fetcher interface {
	Fetch() (items []*gofeed.Item, next time.Time, err error)
}

func Fetch(domain string) Fetcher {
	return &RSSClient{
		channel: domain,
		items:   make([]*gofeed.Item, 0),
		parser:  gofeed.NewParser(),
	}
}

type RSSClient struct {
	channel string
	items   []*gofeed.Item
	parser  *gofeed.Parser
}

func (r *RSSClient) Fetch() (items []*gofeed.Item, next time.Time, err error) {
	//log.Printf("%s start fetching ...\n", r.channel)
	now := time.Now()
	next = now.Add(time.Duration(rand.Intn(5)) * 500 * time.Millisecond)
	result, err := r.parser.ParseURL(r.channel)
	//log.Printf("%v", result.Items)
	if result == nil {
		return
	}
	return result.Items, next, err
}
