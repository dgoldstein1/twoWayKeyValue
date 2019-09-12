package main

import (
	badger "github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"os"
	"strconv"
	"testing"
)

func _WriteEntryHelper(k2v *badger.DB, v2k *badger.DB, s string) error {
	v := rand.Intn(INT_MAX)
	k2v.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(s), []byte(strconv.Itoa(v)))
		return err
	})
	v2k.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.Itoa(v)), []byte(s))
		return err
	})
	return nil
}

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

	t.Run("loads db if already exists", func(t *testing.T) {
		loadPath := "/tmp/twowaykv/iotest/" + strconv.Itoa(rand.Intn(INT_MAX))
		err := os.MkdirAll(loadPath, os.ModePerm)
		require.NoError(t, err)
		defer os.RemoveAll(loadPath)
		// create temp databases in random new dir
		os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
		k2v, v2k, err := ConnectToDb()
		require.Nil(t, err)
		require.NotNil(t, k2v)
		require.NotNil(t, v2k)
		lsm, _ := k2v.Size()
		assert.Equal(t, lsm >= 0, true)
		// write an entry
		testKey := []byte("testingKey")
		testVal := []byte("testingValue")
		err = k2v.Update(func(txn *badger.Txn) error {
			return txn.Set(testKey, testVal)
		})
		require.Nil(t, err)
		err = v2k.Update(func(txn *badger.Txn) error {
			return txn.Set(testVal, testKey)
		})
		require.Nil(t, err)
		// close and reopen
		k2v.Close()
		v2k.Close()
		k2v, v2k, err = ConnectToDb()
		require.Nil(t, err)
		require.NotNil(t, k2v)
		require.NotNil(t, v2k)
		// make sure entries are still there
		err = k2v.View(func(txn *badger.Txn) error {
			item, err := txn.Get(testKey)
			require.Nil(t, err)
			item.Value(func(v []byte) error {
				assert.Equal(t, testVal, v)
				return nil
			})
			return err
		})
		require.Nil(t, err)
		err = v2k.View(func(txn *badger.Txn) error {
			item, err := txn.Get(testVal)
			require.Nil(t, err)
			item.Value(func(k []byte) error {
				assert.Equal(t, testKey, k)
				return nil
			})
			return err
		})
		require.Nil(t, err)
	})
}

func TestGenerateEntry(t *testing.T) {
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

	type Test struct {
		Name             string
		K                string
		ExpectedEntryKey string
		ExpectedError    string
		Setup            func()
		TearDown         func()
	}
	testTable := []Test{
		Test{
			Name:             "generates new Entry succesfully",
			K:                "New Entry",
			ExpectedEntryKey: "New Entry",
			ExpectedError:    "",
			Setup:            func() {},
			TearDown:         func() {},
		},
		Test{
			Name:             "throws error on many collisions",
			K:                "collision",
			ExpectedEntryKey: "collision",
			ExpectedError:    "Too many collisions on creating collision",
			Setup: func() {
				INT_MAX = 1
				_WriteEntryHelper(k2v, v2k, "collision-before")
			},
			TearDown: func() {
				INT_MAX = 9999999
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Setup()
			e, err := GenerateEntry(v2k, test.K)
			assert.Equal(t, test.ExpectedEntryKey, e.Key)
			if err == nil {
				assert.Equal(t, test.ExpectedError, "")
			} else {
				assert.Equal(t, test.ExpectedError, err.Error())
			}
			test.TearDown()
		})
	}
}

func TestCreateIfDoesntExist(t *testing.T) {
	loadPath := "/tmp/twowaykv/" + strconv.Itoa(rand.Intn(INT_MAX))
	err := os.MkdirAll(loadPath, os.ModePerm)
	defer os.RemoveAll(loadPath)
	require.NoError(t, err)
	// setup, create DBs
	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
	k2v, v2k, err := ConnectToDb()
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err)
	assert.NotNil(t, k2v, v2k)
	defer k2v.Close()
	defer v2k.Close()

	type Test struct {
		Name                  string
		Keys                  []string
		MuteAlreadyExists     bool
		ExpectedEntriesLength int
		ExpectedErrors        []string
		Setup                 func()
	}

	testTable := []Test{
		Test{
			Name:                  "adds entries succesfully",
			Keys:                  []string{"test1", "test2"},
			MuteAlreadyExists:     false,
			ExpectedEntriesLength: 2,
			ExpectedErrors:        []string{},
			Setup:                 func() {},
		},
		Test{
			Name:                  "(MuteAlreadyExists=true)",
			Keys:                  []string{"alreadyExists"},
			MuteAlreadyExists:     true,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Setup: func() {
				_WriteEntryHelper(k2v, v2k, "alreadyExists")
			},
		},
		Test{
			Name:                  "(MuteAlreadyExists=false)",
			Keys:                  []string{"alreadyExists1"},
			MuteAlreadyExists:     false,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{"Key alreadyExists1 already exists in DB"},
			Setup: func() {
				_WriteEntryHelper(k2v, v2k, "alreadyExists1")
			},
		},
		Test{
			Name:                  "Mix of already exists and new",
			Keys:                  []string{"key", "key1", "key2", "alreadyExists2"},
			MuteAlreadyExists:     true,
			ExpectedEntriesLength: 4,
			ExpectedErrors:        []string{},
			Setup: func() {
				_WriteEntryHelper(k2v, v2k, "alreadyExists2")
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Setup()
			entries, errors := CreateIfDoesntExist(
				test.Keys,
				test.MuteAlreadyExists,
				k2v,
				v2k,
			)
			assert.Equal(t, test.ExpectedEntriesLength, len(entries))
			assert.Equal(t, test.ExpectedErrors, errors)
		})
	}

}

func TestwriteEntryToDB(t *testing.T) {
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
	k2vWB := k2v.NewWriteBatch()
	v2kWB := v2k.NewWriteBatch()
	defer k2vWB.Cancel()
	defer v2kWB.Cancel()

	type Test struct {
		Name          string
		Key           string
		ExpectedError string
		Setup         func()
		TearDown      func()
	}

	testTable := []Test{
		Test{
			Name:          "creates entry succesfully",
			Key:           "testkey910",
			ExpectedError: "",
			Setup:         func() {},
			TearDown:      func() {},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Setup()
			_, err := writeEntryToDB(v2k, k2vWB, v2kWB, test.Key)
			if err == nil {
				assert.Equal(t, test.ExpectedError, "")
			} else {
				assert.Equal(t, test.ExpectedError, err.Error())
			}
			test.TearDown()
		})
	}

	v2kWB.Flush()
	k2vWB.Flush()

}

func TestReadRandomEntries(t *testing.T) {
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

	type Test struct {
		Name                  string
		n                     int
		ExpectedEntriesLength int
		ExpectedErrorsLength  int
		Setup                 func()
		TearDown              func()
	}

	testTable := []Test{
		Test{
			Name:                  "get one random entry",
			n:                     1,
			ExpectedEntriesLength: 1,
			ExpectedErrorsLength:  0,
			Setup: func() {
				err := v2k.Update(func(txn *badger.Txn) error {
					return txn.Set([]byte(strconv.Itoa(1)), []byte("TEST-KEY"))
				})
				require.Nil(t, err)
			},
			TearDown: func() {
				err := v2k.Update(func(txn *badger.Txn) error {
					return txn.Delete([]byte(strconv.Itoa(1)))
				})
				require.Nil(t, err)
			},
		},
		Test{

			Name:                  "get 5 random entries",
			n:                     5,
			ExpectedEntriesLength: 6,
			ExpectedErrorsLength:  0,
			Setup: func() {
				err := v2k.Update(func(txn *badger.Txn) error {
					for i := 0; i < 5; i++ {
						if e := txn.Set([]byte(strconv.Itoa(i+2)), []byte("TEST-KEY")); e != nil {
							return e
						}
					}
					return nil
				})
				require.Nil(t, err)
			},
			TearDown: func() {

				err := v2k.Update(func(txn *badger.Txn) error {
					for i := 0; i < 5; i++ {
						if e := txn.Delete([]byte(strconv.Itoa(i + 2))); e != nil {
							return e
						}
					}
					return nil
				})
				require.Nil(t, err)
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			test.Setup()
			entries, errors := readRandomEntries(v2k, test.n)
			assert.Equal(t, test.ExpectedEntriesLength, len(entries))
			assert.Equal(t, test.ExpectedErrorsLength, len(errors))
			test.TearDown()
		})
	}
}
