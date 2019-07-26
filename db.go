package main

import (
	"errors"
	"fmt"
	badger "github.com/dgraph-io/badger"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
)

const V2K_PATH = "/v2k"
const K2V_PATH = "/k2v"

// connects to both keyToValue and valueToKey store
func ConnectToDb() (*badger.DB, *badger.DB, error) {
	dir := os.Getenv("GRAPH_DB_STORE_DIR")
	v2kPath := dir + V2K_PATH
	k2vPath := dir + K2V_PATH

	// setup db properties
	options := badger.Options{
		Dir:                     k2vPath,
		ValueDir:                k2vPath,
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
	k2v, err := badger.Open(options)
	if err != nil {
		return nil, nil, err
	}
	// create values => keys DB
	options.Dir = v2kPath
	options.ValueDir = v2kPath
	v2k, err := badger.Open(options)
	return k2v, v2k, err
}

var KEY_NOT_FOUND = "Key not found"
var INT_MAX = 9223372036854775807 // python max int

// writes entry to both dbs
func WriteEntry(k2v *badger.DB, v2k *badger.DB, k string) (Entry, error) {
	v := rand.Intn(INT_MAX)
	val := []byte(strconv.Itoa(v))
	key := []byte(k)
	// assert that keys and values do not already exist
	err := v2k.View(func(txn *badger.Txn) error {
		// keep creating random ints until is found
		keyIsUnique := false
		for !keyIsUnique {
			_, err := txn.Get(val)
			// key not found, stopping condition
			if err != nil && err.Error() == KEY_NOT_FOUND {
				keyIsUnique = true
			} else if err != nil {
				// normal error
				return err
			}
			// key is found somewhere without error, find a new one
			v = rand.Intn(INT_MAX)
			val = []byte(strconv.Itoa(v))
		}
		return nil
	})
	// check that key doesnt exist
	err = k2v.View(func(txn *badger.Txn) error {
		_, err := txn.Get(key)
		// expected KEY_NOT_FOUND error
		if err == nil {
			return fmt.Errorf("Entry already exists in DB for '%s'", k)
		}
		// normal error
		if err.Error() != KEY_NOT_FOUND {
			return err
		}
		// KEY_NOT_FOUND error thrown
		return nil
	})

	// error on retrieval
	if err != nil {
		return Entry{}, err
	}
	// update k2v with k : v
	err = k2v.Update(func(txn *badger.Txn) error {
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

// creates zip file of
func ZipDb() (fileName string, err error) {
	dir := os.Getenv("GRAPH_DB_STORE_DIR")
	fileName = dir + "/twowaykv_export.zip"
	// run zip command in bash
	out, err := exec.Command(
		"zip",
		"-r",
		fileName,
		dir+K2V_PATH,
		dir+V2K_PATH,
	).Output()
	if err != nil {
		err = errors.New(string(out))
	}
	return fileName, err
}
