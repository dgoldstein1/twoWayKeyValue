package server

import (
	"bytes"
	badger "github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateEntry(t *testing.T) {
	t.Run("fails when both are noneType", func(t *testing.T) {
		err := ValidatEntry(Entry{})
		assert.NotNil(t, err)
	})
	t.Run("fails when value is <= 0", func(t *testing.T) {
		err := ValidatEntry(Entry{"test", -3})
		assert.NotNil(t, err)
	})
	t.Run("passes with valid entry", func(t *testing.T) {
		err := ValidatEntry(Entry{"test", 2534})
		assert.Nil(t, err)
	})
}

func TestRetrieveEntry(t *testing.T) {
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	router, s := SetupRouter("*")

	// mock out s.GetEntries DB calls
	testKey := "testKey"
	testval := "2523423426"
	s.GetEntries = func(db *badger.DB, dbKeys []string) (map[string]string, []error) {
		e := map[string]string{}
		e[testKey] = testval
		return e, []error{}
	}

	t.Run("correctly retrieves valid entry", func(t *testing.T) {
		w := httptest.NewRecorder()
		jsonStr := []byte(`[{key : "testKey"}]`)
		req, _ := http.NewRequest("POST", "/entries", bytes.NewBuffer(jsonStr))
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "", w.Body.String())
	})
}
