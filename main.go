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
	Content      string `json:"content"`
	Version      int64  `json:"version"`
	lastAccessed time.Time
}

type Cache struct {
	data   map[string]*Data
	mu     sync.RWMutex
	ticker *time.Ticker
}

// TODO: We don't need the full request, path is enough.
func (c *Cache) Get(r *http.Request) *Data {
	path := r.URL.Path
	c.mu.Lock()
	defer c.mu.Unlock()
	d, ok := c.data[path]
	if !ok {
		d = &Data{}
		c.data[path] = d
	}
	d.lastAccessed = time.Now()
	return d
}

func (c *Cache) Set(r *http.Request, data *Data) {
	path := r.URL.Path
	c.mu.Lock()
	defer c.mu.Unlock()
	data.lastAccessed = time.Now()
	c.data[path] = data
}

func (c *Cache) tidy() {
	c.mu.Lock()
	defer c.mu.Unlock()
	toDelete := make([]string, 0)
	for k, d := range c.data {
		if time.Since(d.lastAccessed) > cacheTimeout {
			toDelete = append(toDelete, k)
		}
	}
	for _, k := range toDelete {
		delete(c.data, k)
	}
}

func NewCache() *Cache {
	c := &Cache{
		data:   make(map[string]*Data),
		ticker: time.NewTicker(cacheCheckInterval),
	}
	// TODO close ticker.
	go func() {
		for range c.ticker.C {
			c.tidy()
		}
	}()
	return c
}

const (
	cacheTimeout       = time.Hour * 24
	cacheCheckInterval = time.Hour
	maxSize            = 1 << 20
	defaultPort        = 7962
)

var (
	//go:embed index.html
	indexDotHTML   []byte
	cache          = NewCache()
	reMaybeBrowser = regexp.MustCompile(`Mozilla|Firefox|Chrome|Safari`)
)

func handleSync(w http.ResponseWriter, r *http.Request) {
	if !isBrowser(r) {
		b, err := ioutil.ReadAll(io.LimitReader(r.Body, maxSize))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		d := cache.Get(r)
		d.Content = string(b)
		d.Version = time.Now().Unix()
		cache.Set(r, d)
		return
	}

	var dd Data
	if err := json.NewDecoder(r.Body).Decode(&dd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	d := cache.Get(r)
	if dd.Version < d.Version {
		// client is too old, send what we have.
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(d)
	} else if dd.Version == d.Version {
		// versions match. nothing to do.
		if d.Version == 0 {
			cache.Set(r, &dd)
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		// update content.
		w.WriteHeader(http.StatusNoContent)
		cache.Set(r, &dd)
	}
}
func isBrowser(r *http.Request) bool {
	return reMaybeBrowser.MatchString(r.Header.Get(`User-Agent`))
}

func indexHTML(w http.ResponseWriter, r *http.Request) {
	if !isBrowser(r) {
		s := cache.Get(r).Content
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
	p := flag.Int(`p`, defaultPort, `port`)
	flag.Parse()
	http.HandleFunc(`/`, index)
	addr := fmt.Sprintf(`0.0.0.0:%d`, *p)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
