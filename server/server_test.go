package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRetrieveEntry(t *testing.T) {
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	router, s := SetupRouter("*")
	// insert data into db
	e := Entry{
		Key:   "k",
		Value: 15,
	}
	WriteEntry(s.valuesToKeys, s.valuesToKeys, e)
	t.Run("correctly retrieves valid entry", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/entry?key=v&value=25", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "", w.Body.String())
	})
	t.Run("fails on invalid int", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/entry?&value=-35", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, 400, w.Code)
		assert.Equal(t, "{\"Code\":400,\"Error\":\"Invalid int '-35' passed on lookup\"}", w.Body.String())
	})
	t.Run("fails when no key or value provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/entry", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, "{\"Code\":400,\"Error\":\"Must provide valid key or value query string\"}", w.Body.String())
	})
	t.Run("fails when entry does not exist", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/entry?key=doesnotexistvalue", nil)
		router.ServeHTTP(w, req)
		// TODO:
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "", w.Body.String())
	})

}
