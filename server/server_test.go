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

	type Test struct {
		Name             string
		Path             string
		Body             []byte
		ExpectedCode     int
		ExpectedResponse string
		Method           string
	}

	testTable := []Test{
		Test{
			Name:             "correctly retrieves valid entry",
			Path:             "/entries",
			Body:             []byte(`[{key : "testKey"}]`),
			ExpectedCode:     200,
			ExpectedResponse: "",
			Method:           "POST",
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)
			assert.Equal(t, test.ExpectedResponse, w.Body.String())
		})

	}

}
