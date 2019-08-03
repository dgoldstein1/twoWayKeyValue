package main

import (
	badger "github.com/dgraph-io/badger"
)

// server environment
type Server struct {
	K2v                 *badger.DB
	V2k                 *badger.DB
	CreateIfDoesntExist func([]string, bool, *badger.DB, *badger.DB) ([]Entry, []string)
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

// util struct for GetEntries
type RetrievalError struct {
	LookupId string // id passed to lookup in DB (either key or value)
	Error    string // error on lookup
	NotFound bool   // is the error that it wasn't found?

}
