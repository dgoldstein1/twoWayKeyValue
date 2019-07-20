package server

import (
	badger "github.com/dgraph-io/badger"
)

// server environment
type Server struct {
	Kv2        *badger.DB
	V2k        *badger.DB
	WriteEntry func(*badger.DB, *badger.DB, Entry) error
	GetEntries func(*badger.DB, []string) (map[string]string, []error)
}

type Entry struct {
	Key   string
	Value int
}

type Error struct {
	Code  int
	Error string
}
