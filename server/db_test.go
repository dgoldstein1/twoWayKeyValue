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
		t.FailNow()
	}
	assert.Nil(t, err)
	assert.NotNil(t, k2v, v2k)
	defer k2v.Close()
	defer v2k.Close()
	t.Run("writes succesful entry to both DBs", func(t *testing.T) {
		key := "testing"
		val := 999
		entry := Entry{key, val}
		err := WriteEntry(k2v, v2k, entry)
		assert.Nil(t, err)
		t.Run("adds correct entry to k:v", func(t *testing.T) {
			k2v.View(func(txn *badger.Txn) error {
				item, _ := txn.Get([]byte(key))
				assert.NotNil(t, item)
				// assert correct key / value
				assert.Equal(t, key, string(item.Key()))
				v, _ := item.Value()
				assert.Equal(t, "999", string(v))
				return nil
			})
		})
		t.Run("adds correct entry to v:k", func(t *testing.T) {
			v2k.View(func(txn *badger.Txn) error {
				val := []byte(strconv.Itoa(val))
				item, _ := txn.Get(val)
				assert.NotNil(t, item)
				// assert correct values
				assert.Equal(t, "999", string(item.Key()))
				v, _ := item.Value()
				assert.Equal(t, key, string(v))
				return nil
			})
		})
		t.Run("does not add to wrong DBs", func(t *testing.T) {
			k2v.View(func(txn *badger.Txn) error {
				val := []byte(strconv.Itoa(val))
				item, _ := txn.Get(val)
				assert.Nil(t, item)
				return nil
			})
			v2k.View(func(txn *badger.Txn) error {
				item, _ := txn.Get([]byte(key))
				assert.Nil(t, item)
				return nil
			})
		})
	})

}
