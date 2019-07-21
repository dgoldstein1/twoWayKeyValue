package server

import (
	badger "github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
)

var testingDir = "/tmp/twowaykv/temp"

func TestConnectToDb(t *testing.T) {
	// setup
	os.MkdirAll(testingDir, os.ModePerm)

	t.Run("succesfully opens both db", func(t *testing.T) {
		os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
		db, db2, err := ConnectToDb()
		assert.Nil(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db2)
		if db != nil {
			lsm, _ := db.Size()
			assert.Equal(t, lsm >= 0, true)
			db.Close()
			db2.Close()
		}
	})
	t.Run("fails on bad db endpoint", func(t *testing.T) {
		os.Setenv("GRAPH_DB_STORE_DIR", "sgfs ;gj2jg////ffk;5")
		db, db2, err := ConnectToDb()
		assert.NotNil(t, err)
		assert.Nil(t, db)
		assert.Nil(t, db2)
	})
}

func TestWriteEntry(t *testing.T) {
	// setup, create DBs
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	k2v, v2k, err := ConnectToDb()
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err)
	assert.NotNil(t, k2v, v2k)
	defer k2v.Close()
	defer v2k.Close()
	t.Run("writes succesful entry to both DBs", func(t *testing.T) {
		t.Run("adds correct entry to k:v", func(t *testing.T) {
			key := "adds correct entry to k:v"
			e, err := WriteEntry(k2v, v2k, key)
			assert.Nil(t, err)
			// lookup in DB
			k2v.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(e.Key))
				assert.Nil(t, err)
				assert.NotNil(t, item)
				// assert correct key / value
				assert.Equal(t, e.Key, string(item.Key()))
				v, _ := item.Value()
				assert.Equal(t, strconv.Itoa(e.Value), string(v))
				return nil
			})
		})
		t.Run("adds correct entry to v:k", func(t *testing.T) {
			key := "adds correct entry to v:k"
			e, err := WriteEntry(k2v, v2k, key)
			assert.Nil(t, err)
			// lookup in DB
			v2k.View(func(txn *badger.Txn) error {
				valAsString := strconv.Itoa(e.Value)
				item, err := txn.Get([]byte(valAsString))
				assert.Nil(t, err)
				assert.NotNil(t, item)
				// assert correct key / value
				assert.Equal(t, valAsString, string(item.Key()))
				k, _ := item.Value()
				assert.Equal(t, e.Key, string(k))
				return nil
			})
		})
		t.Run("does not add if key already exists", func(t *testing.T) {})
	})

}

func TestGetEntries(t *testing.T) {
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	os.MkdirAll(testingDir, os.ModePerm)
	k2v, v2k, err := ConnectToDb()
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err)
	assert.NotNil(t, k2v, v2k)
	defer k2v.Close()
	defer v2k.Close()
	// write entry to DBs
	key := "TESTING_KEY_1"
	valAsString := "999"
	_, err = WriteEntry(k2v, v2k, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("Gets correct entries from string", func(t *testing.T) {
		e, err := GetEntries(k2v, []string{key})
		assert.Equal(t, []RetrievalError{}, err)
		assert.Equal(t, len(e), 1)
		if len(e) == 1 {
			valAsInt, err := strconv.Atoi(e[key])
			assert.Nil(t, err)
			assert.Equal(t, valAsInt < INT_MAX, true)
		}
	})
	t.Run("Gets correct entry from value", func(t *testing.T) {
		e, err := GetEntries(v2k, []string{valAsString})
		assert.Equal(t, []RetrievalError{}, err)
		assert.Equal(t, len(e), 1)
		if len(e) == 1 {
			assert.Equal(t, key, e[valAsString])
		}
	})
	t.Run("returns correct retrieval errors when not found", func(t *testing.T) {
		key := "Sdf23-f2-39if"
		entries, errors := GetEntries(v2k, []string{key})
		assert.Equal(t, 0, len(entries))
		assert.Equal(t, 1, len(errors))
		assert.Equal(t, true, errors[0].NotFound)
		assert.Equal(t, key, errors[0].LookupId)
	})
	t.Run("throws errors on incorrect lookup", func(t *testing.T) {})
}
