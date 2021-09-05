package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
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
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	var dd Data
	if err := json.NewDecoder(r.Body).Decode(&dd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if dd.Version < d.Version {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(d)
	} else if dd.Version == d.Version {
		if d.Version == 0 {
			d = dd
		}
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusNoContent)
		d = dd
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, `index.html`)
}

func main() {
	p := flag.Int(`p`, 7962, `port`)
	flag.Parse()
	http.HandleFunc(`/`, index)
	http.HandleFunc(`/v1/sync`, handleSync)
	http.ListenAndServe(fmt.Sprintf(`0.0.0.0:%d`, *p), nil)
}
