package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Data struct {
	Content string `json:"content"`
	Version int64  `json:"version"`
}

var (
	d  Data
	mu sync.RWMutex
)

func handleSync(w http.ResponseWriter, r *http.Request) {
	if !isBrowser(r) {
		b, err := ioutil.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mu.Lock()
		d.Version = time.Now().Unix()
		d.Content = string(b)
		mu.Unlock()
		return
	}

	var dd Data
	if err := json.NewDecoder(r.Body).Decode(&dd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// json.Decode may last some time on network, don't block.
	mu.Lock()
	defer mu.Unlock()

	if dd.Version < d.Version {
		// client is too old, send what we have.
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(d)
	} else if dd.Version == d.Version {
		// versions match. nothing to do.
		if d.Version == 0 {
			d = dd
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		// update content.
		w.WriteHeader(http.StatusNoContent)
		d = dd
	}
}

//go:embed index.html
var indexDotHTML []byte

var reMaybeBrowser = regexp.MustCompile(`Mozilla|Firefox|Chrome|Safari`)

func isBrowser(r *http.Request) bool {
	return reMaybeBrowser.MatchString(r.Header.Get(`User-Agent`))
}

func indexHTML(w http.ResponseWriter, r *http.Request) {
	if !isBrowser(r) {
		mu.Lock()
		s := d.Content
		mu.Unlock()

		w.Header().Set(`Content-Type`, `text/plain`)
		io.Copy(w, strings.NewReader(s))
		return
	}

	w.Header().Set(`Content-Type`, `text/html; charset=utf-8`)
	http.ServeContent(w, r, ``, time.Now(), bytes.NewReader(indexDotHTML))
}

func index(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		indexHTML(w, r)
		return
	case http.MethodPost:
		handleSync(w, r)
		return
	default:
		http.NotFound(w, r)
	}
}

func main() {
	p := flag.Int(`p`, 7962, `port`)
	flag.Parse()
	http.HandleFunc(`/`, index)
	addr := fmt.Sprintf(`0.0.0.0:%d`, *p)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
