package server

import (
	"bytes"
	"encoding/json"
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
	router, _ := SetupRouter("*")

	// mock out s.GetEntries DB calls
	// testKey := "testKey"
	// testValInt := 2523423426
	// testval := strconv.Itoa(testValInt)

	type Test struct {
		Name                  string
		Path                  string
		Body                  []byte
		ExpectedCode          int
		ExpectedEntriesLength int
		ExpectedErrors        []string
		Method                string
	}

	testTable := []Test{
		Test{
			Name:                  "correctly retrieves valid entry",
			Path:                  "/entries",
			Body:                  []byte(`[{"key":"testKey","value":92238547725307}]`),
			ExpectedCode:          200,
			ExpectedEntriesLength: 1,
			ExpectedErrors:        []string{},
			Method:                "POST",
		},
		// Test{
		// 	Name:             "correctly retrieves valid entry (key only)",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{"key":"testKey"}]`),
		// 	ExpectedCode:     200,
		// 	ExpectedResponse: "{\"errors\":[],\"entries\":[{\"key\":\"testKey\",\"value\":2523423426}]}",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "correctly retrieves valid entry (value only)",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{"value":92236854775807}]`),
		// 	ExpectedCode:     200,
		// 	ExpectedResponse: "",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "adds new key if doesnt exist",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{"key":"testKey2432"}]`),
		// 	ExpectedCode:     200,
		// 	ExpectedResponse: "",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "validates bad int type",
		// 	Path:             "/entries",
		// 	Body:             []byte(`[{"value":"0"}]`),
		// 	ExpectedCode:     400,
		// 	ExpectedResponse: "{\"Code\":400,\"Error\":\"json: cannot unmarshal string into Go struct field Entry.value of type int\"}",
		// 	Method:           "POST",
		// },
		// Test{
		// 	Name:             "bad jsob buffer",
		// 	Path:             "/entries",
		// 	Body:             []byte(`2085jf2 3j0d sdf}`),
		// 	ExpectedCode:     400,
		// 	ExpectedResponse: "{\"Code\":400,\"Error\":\"json: cannot unmarshal number into Go value of type []server.Entry\"}",
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
			resp := RetrieveEntryResponse{}
			body := []byte(w.Body.String())
			err := json.Unmarshal(body, &resp)
			assert.Nil(t, err)
			assert.Equal(t, test.ExpectedEntriesLength, len(resp.Entries))
			assert.Equal(t, test.ExpectedErrors, resp.Errors)
		})

	}

}
