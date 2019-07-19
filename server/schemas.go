package server

import (
	badger "github.com/dgraph-io/badger"
)

// server environment
type Server struct {
	keysToValues *badger.DB
	valuesToKeys *badger.DB
}
