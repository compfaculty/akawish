package main

import (
	"github.com/mmcdole/gofeed"
	"log"
	"sync"
	"time"
)

const (
	maxPendingItems  = 10
	nextTryTimeDelay = 10
)

type Subscribtion interface {
	Updates() <-chan *gofeed.Item
	Close() error
}

func Subscribe(rssClient Fetcher) Subscribtion {
	//log.Printf("Sub added %v\n", rssClient)
	s := &rssSub{
		fetcher: rssClient,
		updates: make(chan *gofeed.Item),
		closing: make(chan chan error),
	}
	go s.loop()
	//log.Printf("return %v\n", s)
	return s
}

type rssSub struct {
	fetcher Fetcher
	updates chan *gofeed.Item
	closing chan chan error
}

func (r *rssSub) Updates() <-chan *gofeed.Item {
	return r.updates
}

func (r *rssSub) Close() error {
	//log.Println("closing...")
	errc := make(chan error)
	r.closing <- errc
	return <-errc
}

func (r *rssSub) loop() {
	//log.Println("loop has started...")
	type fetchedResult struct {
		items []*gofeed.Item
		next  time.Time
		err   error
	}

	var (
		fetchDoneChan chan fetchedResult
		pendingItems  []*gofeed.Item
		currentError  error
		nextTryTime   time.Time
		seenLinksMap  sync.Map
	)

	for {
		var (
			fetchDelay     time.Duration
			currentItem    *gofeed.Item
			downstreamChan chan *gofeed.Item
		)
		if now := time.Now(); nextTryTime.After(now) {
			fetchDelay = nextTryTime.Sub(now)
		}
		var startFetchChan <-chan time.Time
		if fetchDoneChan == nil && len(pendingItems) < maxPendingItems {
			//log.Println("startfetch")
			startFetchChan = time.After(fetchDelay)
		}
		if len(pendingItems) > 0 {
			currentItem = pendingItems[0]
			downstreamChan = r.updates
		}
		select {
		case errc := <-r.closing:
			log.Printf("closing subscribtion %q...\n", r.fetcher.(*RSSClient).channel)
			errc <- currentError
			close(r.updates)
			return
		case <-startFetchChan:
			fetchDoneChan = make(chan fetchedResult, 1)
			go func() {
				fetchedItems, nextTryTime, currentError := r.fetcher.Fetch()
				fr := fetchedResult{
					items: fetchedItems,
					next:  nextTryTime,
					err:   currentError,
				}
				//log.Printf("fetched result %v\n", fr)
				fetchDoneChan <- fr
			}()
		case result := <-fetchDoneChan:
			fetchDoneChan = nil
			fetchedItems := result.items
			nextTryTime, currentError = result.next, result.err
			if currentError != nil {
				log.Printf("error %v , waiting 10 sec\n", currentError)
				nextTryTime = time.Now().Add(nextTryTimeDelay * time.Second)
			}
			for _, item := range fetchedItems {
				if _, ok := seenLinksMap.Load(item.GUID); !ok {
					//log.Printf("add to pendingItemss %v \n", item)
					pendingItems = append(pendingItems, item)
					seenLinksMap.Store(item.GUID, struct{}{})
				}
			}
		case downstreamChan <- currentItem:
			pendingItems = pendingItems[1:]
		}
	}
}
