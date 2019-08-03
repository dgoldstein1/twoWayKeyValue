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

var INT_MAX = 999999999 // python max int

// creates new Entry object to be written
// assumed that key is not duplicate
func GenerateEntry(v2k *badger.DB, k string) (Entry, error) {
	v := rand.Intn(INT_MAX)
	val := []byte(strconv.Itoa(v))
	// assert that keys and values do not already exist
	err := v2k.View(func(txn *badger.Txn) error {
		// keep creating random ints until is found
		keyIsUnique := false
		i := 0
		for !keyIsUnique {
			_, err := txn.Get(val)
			// key not found, stopping condition
			if err != nil && err == badger.ErrKeyNotFound {
				keyIsUnique = true
			} else if err != nil {
				// normal error
				return err
			} else if i == INT_MAX {
				return fmt.Errorf("Too many collisions on creating %s", k)
			}
			// key is found somewhere without error, find a new one
			i++
			v = rand.Intn(INT_MAX)
			val = []byte(strconv.Itoa(v))
		}
		return nil
	})
	return Entry{k, v}, err
}

// adds new entry to DB if doesnt already exist
// MuteAlreadyExists does not add errors to list if key already exists
func CreateIfDoesntExist(
	keys []string,
	muteAlreadyExists bool,
	k2v *badger.DB,
	v2k *badger.DB,
) (
	entries []Entry,
	errors []string,
) {
	// initialize return variables
	entries = []Entry{}
	errors = []string{}
	keysToWriteToDB := []string{}
	// find entries to create
	k2v.View(func(txn *badger.Txn) error {
		for _, k := range keys {
			// expect KEY_NOT_FOUND error
			_, err := txn.Get([]byte(k))
			if err == badger.ErrKeyNotFound {
				keysToWriteToDB = append(keysToWriteToDB, k)
			} else if !muteAlreadyExists && err == nil {
				// key already exists in DB
				errors = append(errors, fmt.Sprintf("Key %s already exists in DB", k))
			} else if err != nil {
				// io error on lookup
				logErr("Error on looking up key %s: %v", k, err)
				errors = append(errors, err.Error())
			}
		}
		return nil
	})

	// batch write keys
	k2vWB := k2v.NewWriteBatch()
	v2kWB := v2k.NewWriteBatch()
	defer k2vWB.Cancel()
	defer v2kWB.Cancel()
	// write entries to both DBs
	for _, k := range keysToWriteToDB {
		e, err := writeEntryToDB(v2k, k2vWB, v2kWB, k)
		if err != nil {
			logErr("Could not create entry %+v: %v", e, err)
		} else {
			entries = append(entries, e)
		}
	}
	// flush transactions
	if err := v2kWB.Flush(); err != nil {
		logErr("Error flushing v2k Write Batch %v", err)
		errors = append(errors, err.Error())
	}
	if err := k2vWB.Flush(); err != nil {
		logErr("Error flushing k2v Write Batch %v", err)
		errors = append(errors, err.Error())
	}
	return entries, errors
}

// creates and writes a new entry to DB in batch mode
func writeEntryToDB(
	v2k *badger.DB,
	kv2WB *badger.WriteBatch,
	v2kWB *badger.WriteBatch,
	key string,
) (e Entry, err error) {
	// create new
	e, err = GenerateEntry(v2k, key)
	if err != nil {
		logErr("Error generating entry %s: %v", key, err)
		return Entry{}, err
	}
	// write to DB
	v := []byte(strconv.Itoa(e.Value))
	k := []byte(e.Key)
	if err = kv2WB.Set(k, v); err != nil {
		logErr("Error setting k2v %+v: %v", e, err)
		return Entry{}, err
	}
	if err = v2kWB.Set(v, k); err != nil {
		logErr("Error setting v2k %+v: %v", e, err)
	}
	return e, err
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
