package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
	"strconv"
)

// mock out logging calls for testing
var logFatalf = log.Fatalf
var logWarn = log.Warnf
var logMsg = log.Infof
var logErr = log.Errorf
var logDebug = log.Debugf

// entrypoint
func SetupRouter(docs string) (*gin.Engine, *Server) {
	// try to connect to db
	keysToValues, valuesToKeys, err := ConnectToDb()
	if err != nil {
		logFatalf("Could not establish connection to db: %v", err)
	}
	// create server object
	s := Server{keysToValues, valuesToKeys}
	// define endpoints
	router := gin.Default()
	router.Use(gin.Logger())
	// set base page as readme html
	router.LoadHTMLGlob(docs)
	router.Static("/static", "static")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	// metrics
	p := ginprometheus.NewPrometheus("gin")
	p.Use(router)
	// core endpoints
	router.GET("/entry", s.RetreieveEntry)
	router.POST("/entry", s.CreateEntry)
	router.GET("/save", s.ExportDB)
	// return server
	return router, &s
}

// validate key and value
func ValidateArgs(key string, value int) error {
	if key == "" && value == 0 {
		return errors.New("Must provide valid key or value query string")
	}
	if key == "" && value <= 0 {
		return fmt.Errorf("Invalid int '%d' passed on lookup", value)
	}
	return nil
}

// retrieve and try from db
func (s *Server) RetreieveEntry(c *gin.Context) {
	key := c.Query("key")
	val, _ := strconv.Atoi(c.Query("value"))
	// valdate args
	err := ValidateArgs(key, val)
	if err != nil {
		c.JSON(400, Error{
			Error: err.Error(),
			Code:  400,
		})
		return
	}

}

// create new entry in db
func (s *Server) CreateEntry(c *gin.Context) {

}

// export db to file
func (s *Server) ExportDB(c *gin.Context) {

}
