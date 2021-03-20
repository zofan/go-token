package token

import (
	"time"
)

type Storage interface {
	Init() error
	Close() error

	Get(id string) (*Token, error)
	Set(*Token) error
}

type Token struct {
	ID string

	Created time.Time
	Expired time.Time

	Access  []string
	Account uint64
}

func (t *Token) IsExpired() bool {
	return time.Now().After(t.Expired)
}

type Metric struct {
	Request uint64
	Success uint64
	Error   uint64
	Last    time.Time
}
