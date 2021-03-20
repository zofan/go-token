package token

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type External struct {
	cache  map[string]*Token
	hits   map[string]time.Time
	config url.Values
	mu     sync.RWMutex
}

func NewExternal(dsn string) *External {
	ts := &External{
		cache: make(map[string]*Token),
		hits:  make(map[string]time.Time),
	}

	ts.config, _ = url.ParseQuery(dsn)

	return ts
}

func (ts *External) Get(id string) (*Token, error) {
	ts.mu.RLock()
	t, ok := ts.cache[id]
	if ok {
		ts.mu.RUnlock()
		ts.mu.Lock()
		ts.hits[id] = time.Now()
		ts.mu.Unlock()
		return t, nil
	}
	ts.mu.RUnlock()

	resp, err := http.Get(ts.config.Get(`url`))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	de := json.NewDecoder(resp.Body)
	err = de.Decode(t)
	if err != nil {
		return t, err
	}

	ts.mu.Lock()
	ts.cache[id] = t
	ts.hits[id] = time.Now()
	ts.mu.Unlock()

	return t, nil
}

func (ts *External) Set(t *Token) error {
	buf := bytes.NewBuffer([]byte{})

	en := json.NewEncoder(buf)
	err := en.Encode(t)
	if err != nil {
		return err
	}

	resp, err := http.Post(ts.config.Get(`url`), `application/json`, buf)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(`fail token store`)
	}

	ts.mu.Lock()
	ts.cache[t.ID] = t
	ts.hits[t.ID] = time.Now()
	ts.mu.Unlock()

	return nil
}

func (ts *External) Init() error {
	go ts.gcWorker()

	return nil
}

func (ts *External) gcWorker() {
	cacheLife, _ := time.ParseDuration(ts.config.Get(`cacheLife`))
	if cacheLife == 0 {
		cacheLife = time.Minute
	}

	for range time.Tick(cacheLife) {
		ts.mu.Lock()
		for id, t := range ts.hits {
			if time.Since(t) > cacheLife {
				delete(ts.cache, id)
				delete(ts.hits, id)
			}
		}
		ts.mu.Unlock()
	}
}

func (ts *External) Close() error {
	return nil
}
