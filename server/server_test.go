package server

import (
	"bytes"
	"errors"
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
		if len(dbKeys) == 1 && dbKeys[0] == testKey {
			return e, []error{}
		}
		// simulate failure
		return e, []error{errors.New("Key not found")}
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
			Body:             []byte(`[{"key":"testKey","value":0}]`),
			ExpectedCode:     200,
			ExpectedResponse: "",
			Method:           "POST",
		},
		// Test{
		// 	Name:             "fails on invalid entry",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{}]`),
		// 	ExpectedCode:     400,
		// 	ExpectedResponse: "{error}",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "fails on db error",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{key : key won't be found}]`),
		// 	ExpectedCode:     400,
		// 	ExpectedResponse: "{error}",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "bad jsob buffer",
		// 	Path:             "/entries",
		// 	Body:             []byte(`2085jf2 3j0d sdf}`),
		// 	ExpectedCode:     400,
		// 	ExpectedResponse: "{error}",
		// 	Method:           "POST",
		// },
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
			req.Header.Add("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			assert.Equal(t, test.ExpectedCode, w.Code)
			assert.Equal(t, test.ExpectedResponse, w.Body.String())
		})

	}

}
