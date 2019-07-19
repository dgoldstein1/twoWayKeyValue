package server

import (
	badger "github.com/dgraph-io/badger"
	"os"
)

// connects to both keyToValue and valueToKey store
func ConnectToDb() (*badger.DB, *badger.DB, error) {
	dir := os.Getenv("GRAPH_DB_STORE_DIR")
	// setup db properties
	options := badger.Options{
		Dir:                     dir + "/keysToValues",
		ValueDir:                dir + "/keysToValues",
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
	// create keys => values DB
	keysToValuesDB, err := badger.Open(options)
	if err != nil {
		return nil, nil, err
	}
	// create values => keys DB
	options.Dir = dir + "/valuesToKeys"
	options.ValueDir = dir + "/valuesToKeys"
	valuesToKeysDB, err := badger.Open(options)
	return keysToValuesDB, valuesToKeysDB, err
}

// write entry to 'key' store
func WriteKey() error {
	return nil
}

// write entry to 'value' store
func WriteValue() error {
	return nil
}
