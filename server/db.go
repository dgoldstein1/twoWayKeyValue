package server

import (
	badger "github.com/dgraph-io/badger"
	"math/rand"
	"os"
	"strconv"
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

var KEY_NOT_FOUND = "Key not found"
var INT_MAX = 9223372036854775807 // python max int

// writes entry to both dbs
func WriteEntry(k2v *badger.DB, v2k *badger.DB, k string) (Entry, error) {
	v := rand.Intn(INT_MAX)
	val := []byte(strconv.Itoa(v))
	// assert that does not already exist

	key := []byte(k)
	// update k2v with k : v
	err := k2v.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
	// write v:k
	err = v2k.Update(func(txn *badger.Txn) error {
		return txn.Set(val, key)
	})
	return Entry{k, v}, err
}

// retrieves entry using either key or value
func GetEntries(db *badger.DB, dbKeys []string) (map[string]string, []RetrievalError) {
	errors := []RetrievalError{}
	entries := map[string]string{}
	// read from DB
	db.View(func(txn *badger.Txn) error {
		// read each key in DB
		for _, k := range dbKeys {
			item, err := txn.Get([]byte(k))
			if err != nil {
				errors = append(errors, RetrievalError{
					LookupId: k,
					Error:    err.Error(),
					NotFound: err.Error() == KEY_NOT_FOUND,
				})
				break
			}
			// key exists
			v, err := item.Value()
			if err != nil {
				errors = append(errors, RetrievalError{k, err.Error(), false})
			}
			// add new Entry to list
			entries[k] = string(v)
		}
		// return out of View function
		return nil
	})
	return entries, errors
}
