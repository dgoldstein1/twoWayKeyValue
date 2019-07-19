package server

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestRetrieveEntry(t *testing.T) {
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	router, s := SetupRouter()
	// insert data into db
	e := Entry{
		Key:   "k",
		Value: 15,
	}
	WriteEntry(s.valuesToKeys, s.valuesToKeys, e)
	gin.SetMode(gin.TestMode)
	t.Run("correctly retrieves valid entry", func(t *testing.T) {

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/entry?key=v&value=25", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "pong", w.Body.String())

	})
	// t.Run("fails on invalid int", func(t *testing.T) {
	// 	c, w := _createContextHelper()
	// 	c.Params = []gin.Param{gin.Param{Value: "2050--3-4593"}}
	// 	s.RetreieveEntry(c)
	// 	assert.Equal(t, w.Code, 400)
	// 	b, _ := ioutil.ReadAll(w.Body)
	// 	if w.Body != nil {
	// 		res := Error{}
	// 		err := json.Unmarshal(b, &res)
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, 400, res.Code)
	// 		assert.Equal(t, "Invalid int", res.Error)
	// 	}
	// })
	// t.Run("fails when no key or value provided", func(t *testing.T) {
	// 	c, w := _createContextHelper()
	// 	c.Params = []gin.Param{gin.Param{}}
	// 	s.RetreieveEntry(c)
	// 	assert.Equal(t, w.Code, 400)
	// 	b, _ := ioutil.ReadAll(w.Body)
	// 	if w.Body != nil {
	// 		res := Error{}
	// 		err := json.Unmarshal(b, &res)
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, 400, res.Code)
	// 		assert.Equal(t, "Bad Request", res.Error)
	// 	}
	// })
	// t.Run("fails when entry does not exist", func(t *testing.T) {
	// 	c, w := _createContextHelper()
	// 	c.Params = []gin.Param{gin.Param{Key: "2523234234f23f23d"}}
	// 	s.RetreieveEntry(c)
	// 	assert.Equal(t, w.Code, 404)
	// 	b, _ := ioutil.ReadAll(w.Body)
	// 	if w.Body != nil {
	// 		res := Error{}
	// 		err := json.Unmarshal(b, &res)
	// 		assert.Nil(t, err)
	// 		assert.Equal(t, 404, res.Code)
	// 		assert.Equal(t, "Entry Not Found", res.Error)
	// 	}
	// })

}
