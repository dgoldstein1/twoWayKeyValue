package server

import (
	badger "github.com/dgraph-io/badger"
	"os"
)

func ConnectToDb() (*badger.DB, error) {
	dir := os.Getenv("GRAPH_DB_STORE_DIR")
	// setup db properties
	options := badger.Options{
		Dir:                     dir,
		ValueDir:                dir,
		LevelOneSize:            256 << 20,
		LevelSizeMultiplier:     10,
		MaxLevels:               7,
		MaxTableSize:            64 << 20,
		NumCompactors:           2, // Compactions can be expensive. Only run 2.
		NumLevelZeroTables:      5,
		NumLevelZeroTablesStall: 10,
		NumMemtables:            5,
		SyncWrites:              true,
		NumVersionsToKeep:       1,
		ValueLogFileSize:        1<<30 - 1,
		ValueLogMaxEntries:      1000000,
		ValueThreshold:          32,
		Truncate:                false,
	}
	return badger.Open(options)
}
