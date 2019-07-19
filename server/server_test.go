package server

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"
)

func _createContextHelper() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestRetrieveEntry(t *testing.T) {
	os.Setenv("GRAPH_DB_STORE_DIR", testingDir)
	db, db2, _ := ConnectToDb()
	s := Server{db, db2}
	// insert data into db
	e := Entry{
		Key:   "k",
		Value: 15,
	}
	WriteEntry(db, db2, e)
	c, w := _createContextHelper()
	c.Params = []gin.Param{gin.Param{Key: e.Key, Value: "15"}}
	s.RetreieveEntry(c)
	assert.Equal(t, w.Code, 200)
	b, _ := ioutil.ReadAll(w.Body)
	if w.Body != nil {
		res := Entry{}
		err := json.Unmarshal(b, &res)
		assert.Nil(t, err)
		assert.Equal(t, e.Key, res.Key)
		assert.Equal(t, e.Value, res.Value)
	}

}
