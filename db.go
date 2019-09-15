package main

import (
	"fmt"
	badger "github.com/dgraph-io/badger"
	"math/rand"
	"os"
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
	options := badger.DefaultOptions(k2vPath)
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
			item, err := txn.Get([]byte(k))
			if err == badger.ErrKeyNotFound {
				keysToWriteToDB = append(keysToWriteToDB, k)
			} else if err == nil {
				// key already exists in DB
				if !muteAlreadyExists {
					errors = append(errors, fmt.Sprintf("Key %s already exists in DB", k))
				}
				// add to response
				key := string(item.KeyCopy(nil))
				v, _ := item.ValueCopy(nil)
				val, _ := strconv.Atoi(string(v))
				entries = append(entries, Entry{key, val})
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

// reads a number of random entries from DB
func readRandomEntries(
	v2k *badger.DB,
	n int,
) (
	entries []Entry,
	err error,
) {
	// open up DB read
	err = v2k.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = n
		it := txn.NewIterator(opts)
		defer it.Close()
		// keep track of tries
		m := make(map[int]bool)
		maxRetries := n * 5
		tries := 0
		// loop through different random numbers and seek at that n
		for prefix := rand.Intn(INT_MAX); len(entries) < n; prefix = rand.Intn(INT_MAX) {
			// start iterator at random N
			it.Seek([]byte(strconv.Itoa(prefix)))
			if it.Valid() {
				k, _ := strconv.Atoi(string(it.Item().Key()))
				// prefix found, make sure id is't 0
				if k != 0 && !m[k] {
					it.Item().Value(func(v []byte) error {
						// add to entries
						entries = append(entries, Entry{string(v), k})
						m[k] = true
						return nil
					})
				}
			}
			// could not find key, incr tries
			tries++
			// return error if too many retries
			if tries > maxRetries {
				return fmt.Errorf("max collisions reached finding random entries")
			}
		}
		// exit db.view
		return nil
	})
	return entries, err
}

func GetEntriesFromKeys(k2v *badger.DB, keys []string) (entries []Entry, errors []string) {
	return entries, errors
}
