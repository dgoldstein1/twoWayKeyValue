package main

//
// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"math/rand"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"strconv"
// 	"testing"
// )
//
// func TestValidateEntry(t *testing.T) {
// 	t.Run("fails when both are noneType", func(t *testing.T) {
// 		err := ValidatEntry(Entry{})
// 		assert.NotNil(t, err)
// 	})
// 	t.Run("fails when value is <= 0", func(t *testing.T) {
// 		err := ValidatEntry(Entry{"test", -3})
// 		assert.NotNil(t, err)
// 	})
// 	t.Run("passes with valid entry", func(t *testing.T) {
// 		err := ValidatEntry(Entry{"test", 2534})
// 		assert.Nil(t, err)
// 	})
// }
//
// func TestRetrieveEntry(t *testing.T) {
// 	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
// 	router, _ := SetupRouter("./api/*")
//
// 	type Test struct {
// 		Name                  string
// 		Path                  string
// 		Body                  []byte
// 		ExpectedCode          int
// 		ExpectedEntriesLength int
// 		ExpectedErrors        []string
// 		Method                string
// 	}
// 	// used for testing valid value lookup
// 	validTestValue := ""
//
// 	testTable := []Test{
// 		Test{
// 			Name:                  "correctly retrieves valid entry",
// 			Path:                  "/entries",
// 			Body:                  []byte(`[{"key":"testKey","value":92238547725307}]`),
// 			ExpectedCode:          200,
// 			ExpectedEntriesLength: 1,
// 			ExpectedErrors:        []string{},
// 			Method:                "POST",
// 		},
// 		Test{
// 			Name:                  "correctly retrieves valid entry (key only)",
// 			Path:                  "/entries",
// 			Body:                  []byte(`[{"key":"testKey"}]`),
// 			ExpectedCode:          200,
// 			ExpectedEntriesLength: 1,
// 			ExpectedErrors:        []string{},
// 			Method:                "POST",
// 		},
// 		Test{
// 			Name:                  "correctly retrieves valid entry (value only)",
// 			Path:                  "/entries",
// 			Body:                  []byte(""), // set at execution time
// 			ExpectedCode:          200,
// 			ExpectedEntriesLength: 1,
// 			ExpectedErrors:        []string{},
// 			Method:                "POST",
// 		},
// 		Test{
// 			Name:                  "adds new key if doesnt exist",
// 			Path:                  "/entries",
// 			Body:                  []byte(`[{"key":"testKey2432"}]`),
// 			ExpectedCode:          200,
// 			ExpectedEntriesLength: 1,
// 			ExpectedErrors:        []string{},
// 			Method:                "POST",
// 		},
// 		Test{
// 			Name:                  "validates bad int type",
// 			Path:                  "/entries",
// 			Body:                  []byte(`[{"value":"0"}]`),
// 			ExpectedCode:          400,
// 			ExpectedEntriesLength: 0,
// 			ExpectedErrors:        []string{"json: cannot unmarshal string into Go struct field Entry.value of type int"},
// 			Method:                "POST",
// 		},
// 		Test{
// 			Name:                  "bad json buffer",
// 			Path:                  "/entries",
// 			Body:                  []byte(`2085jf2 3j0d sdf}`),
// 			ExpectedCode:          400,
// 			ExpectedEntriesLength: 0,
// 			ExpectedErrors:        []string{"json: cannot unmarshal number into Go value of type []main.Entry"},
// 			Method:                "POST",
// 		},
// 	}
//
// 	for _, test := range testTable {
// 		t.Run(test.Name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			if test.Name == "correctly retrieves valid entry (value only)" {
// 				test.Body = []byte(`[{"value" : ` + string(validTestValue) + `}]`)
// 			}
// 			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
// 			req.Header.Add("Content-Type", "application/json")
// 			router.ServeHTTP(w, req)
// 			assert.Equal(t, test.ExpectedCode, w.Code)
//
// 			fmt.Println(" ****> POST: " + string(test.Body))
// 			body := []byte(w.Body.String())
// 			if test.ExpectedCode == 200 {
// 				resp := RetrieveEntryResponse{}
// 				err := json.Unmarshal(body, &resp)
// 				assert.Nil(t, err)
// 				assert.Equal(t, test.ExpectedEntriesLength, len(resp.Entries))
// 				assert.Equal(t, test.ExpectedErrors, resp.Errors)
// 				// set createdEntry on success
// 				if len(resp.Entries) > 0 {
// 					assert.NotEqual(t, 0, resp.Entries[0].Value)
// 					validTestValue = strconv.Itoa(resp.Entries[0].Value)
// 					assert.NotEqual(t, "0", validTestValue)
// 				}
// 			} else {
// 				resp := Error{}
// 				err := json.Unmarshal(body, &resp)
// 				assert.Nil(t, err)
// 				assert.Equal(t, test.ExpectedErrors[0], resp.Error)
// 				assert.Equal(t, test.ExpectedCode, resp.Code)
// 			}
// 		})
//
// 	}
// }
//
// func TestExportDb(t *testing.T) {
// 	loadPath := "/tmp/twowaykv/iotest/" + strconv.Itoa(rand.Intn(INT_MAX))
// 	err := os.MkdirAll(loadPath, os.ModePerm)
// 	require.NoError(t, err)
// 	defer os.RemoveAll(loadPath)
// 	os.Setenv("GRAPH_DB_STORE_DIR", loadPath)
// 	router, _ := SetupRouter("./api/*")
//
// 	type Test struct {
// 		Name           string
// 		Path           string
// 		Body           []byte
// 		ExpectedCode   int
// 		ExpectedErrors []string
// 		Method         string
// 	}
// 	// used for testing valid value lookup
//
// 	testTable := []Test{
// 		Test{
// 			Name:           "returns error if file could not be found",
// 			Path:           "/save",
// 			Body:           []byte(""),
// 			ExpectedCode:   500,
// 			ExpectedErrors: []string{"Not implemented"},
// 			Method:         "GET",
// 		},
// 	}
//
// 	for _, test := range testTable {
// 		t.Run(test.Name, func(t *testing.T) {
// 			w := httptest.NewRecorder()
// 			req, _ := http.NewRequest(test.Method, test.Path, bytes.NewBuffer(test.Body))
// 			req.Header.Add("Content-Type", "application/json")
// 			router.ServeHTTP(w, req)
// 			assert.Equal(t, test.ExpectedCode, w.Code)
//
// 			fmt.Println(" ****> GET: " + test.Path)
// 			body := []byte(w.Body.String())
// 			if test.ExpectedCode == 200 {
// 				// TODO
// 			} else {
// 				resp := Error{}
// 				err := json.Unmarshal(body, &resp)
// 				assert.Nil(t, err)
// 				assert.Equal(t, test.ExpectedErrors[0], resp.Error)
// 				assert.Equal(t, test.ExpectedCode, resp.Code)
// 			}
// 		})
//
// 	}
// }
