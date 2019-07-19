package server

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestConnectToDb(t *testing.T) {
	// setup
	testingDir := "/tmp/twowaykv/temp"
	os.MkdirAll(testingDir, os.ModePerm)

	t.Run("succesfully opens both db", func(t *testing.T) {
		os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
		db, db2, err := ConnectToDb()
		assert.Nil(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, db2)
		if db != nil {
			lsm, _ := db.Size()
			assert.Equal(t, lsm, int64(0))
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
