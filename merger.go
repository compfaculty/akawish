package main

import (
	"errors"
	"github.com/mmcdole/gofeed"
	"strings"
)

type merger struct {
	subs    []Subscribtion
	updates chan *gofeed.Item
	exit    chan struct{}
	errs    chan error
}

func (m *merger) Updates() <-chan *gofeed.Item {
	return m.updates
}

func (m merger) Close() (err error) {
	close(m.exit)
	var errLog strings.Builder
	fail := false
	for range m.subs {
		if e := <-m.errs; e != nil {
			errLog.WriteString(e.Error() + ";")
			fail = true
		}
	}
	close(m.updates)
	if fail {
		return errors.New(errLog.String())
	}
	return nil
}

func Merge(subs ...Subscribtion) Subscribtion {
	m := &merger{
		subs:    subs,
		updates: make(chan *gofeed.Item, 0),
		exit:    make(chan struct{}, 0),
		errs:    make(chan error, 0),
	}
	for _, sub := range subs {
		//log.Printf("started fetch from sub %v\n", sub)
		go func(s Subscribtion) {
			for {
				var item *gofeed.Item
				select {
				case item = <-s.Updates():
					//log.Printf("got item %v\n", item)
				case <-m.exit:
					m.errs <- s.Close()
					return
				}
				select {
				case m.updates <- item:
				case <-m.exit:
					m.errs <- s.Close()
					return
				}
			}
		}(sub)
	}
	return m
}
