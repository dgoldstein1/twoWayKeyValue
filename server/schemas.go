package server

import (
	badger "github.com/dgraph-io/badger"
)

// server environment
type Server struct {
	K2v        *badger.DB
	V2k        *badger.DB
	WriteEntry func(*badger.DB, *badger.DB, Entry) error
	GetEntries func(*badger.DB, []string) (map[string]string, []error)
}

type Entry struct {
	Key   string `json:"key" binding:"required"`
	Value int    `json:"value" binding:"required"`
}

type RetrieveEntryResponse struct {
	Errors  []string `json:"errors" binding:"required"`
	Entries []Entry  `json:"entries" binding:"required"`
}

type Error struct {
	Code  int
	Error string
}
