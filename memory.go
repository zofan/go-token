package token

import (
	"github.com/zofan/go-fread"
	"github.com/zofan/go-fwrite"
	"net/url"
	"sync"
)

type InMemory struct {
	storage map[string]*Token
	config  url.Values
	mu      sync.RWMutex
}

func NewInMemory(dsn string) *InMemory {
	ts := &InMemory{
		storage: make(map[string]*Token),
	}

	ts.config, _ = url.ParseQuery(dsn)

	return ts
}

func (ts *InMemory) Get(id string) (*Token, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	t, ok := ts.storage[id]
	if !ok {
		return nil, ErrTokenNotFound
	}

	return t, nil
}

func (ts *InMemory) Set(t *Token) error {
	ts.mu.Lock()
	ts.storage[t.ID] = t
	ts.mu.Unlock()

	return ts.persist()
}

func (ts *InMemory) Init() error {
	return fread.ReadJson(ts.config.Get(`file`), &ts.storage)
}

func (ts *InMemory) Close() error {
	return ts.persist()
}

func (ts *InMemory) persist() error {
	return fwrite.WriteJson(ts.config.Get(`file`), ts.storage)
}
