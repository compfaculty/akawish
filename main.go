package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", "127.0.0.1:8888", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func main() {
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	merged := Merge(
		Subscribe(Fetch("http://feeds.feedburner.com/adb_news?format=xml")),
		Subscribe(Fetch("https://rss.art19.com/apology-line")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_world.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_europe.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_meast.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_us.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_sport.rss")),
		Subscribe(Fetch("http://rss.cnn.com/rss/edition_space.rss")),
	)

	go func() {
		//for feed := range merged.Updates(){
		//	hub.broadcast <- []byte(feed.Title)
		//}
		tick := time.Tick(1 * time.Second)
		for {
			select {
			case <-tick:
				d := <-merged.Updates()
				if d != nil && d.Title != "" {
					hub.broadcast <- []byte(d.Title)
				}
			}
		}
	}()
	log.Println("Server started...")
	log.Fatal(http.ListenAndServe("localhost:8888", nil))
}
