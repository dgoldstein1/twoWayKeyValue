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
	kDB, vDB, err := ConnectToDb()
	if err != nil {
		logFatalf("Could not establish connection to db: %v", err)
	}
	// create server object
	s := Server{kDB, vDB, WriteEntry, GetEntries}
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
	router.POST("/entries", s.RetreieveEntries)
	router.GET("/save", s.ExportDB)
	// return server
	return router, &s
}

// validate key and value
func ValidatEntry(e Entry) error {
	if e.Key == "" && e.Value == 0 {
		return errors.New("Must provide valid key or value query string")
	}
	if e.Value <= 0 {
		return fmt.Errorf("Invalid int '%d'", e.Value)
	}
	return nil
}

// retrieve and try from db
func (s *Server) RetreieveEntries(c *gin.Context) {
	// read in request
	entriesPassed := []Entry{}
	if err := c.BindJSON(&entriesPassed); err != nil {
		c.JSON(400, Error{400, err.Error()})
		return
	}
	if len(entriesPassed) == 0 {
		c.JSON(400, "Bad []entry or no entries passed")
		return
	}
	// create big array entries for keys and values
	// entriesToReturn := []Entry{}
	k2vToFetch := []string{}
	v2kToFetch := []string{}
	for _, e := range entriesPassed {
		// validate entry
		if err := ValidatEntry(e); err != nil {
			c.JSON(400, err.Error())
			return
		}
		// if has key, lookup by key, else lookup by value
		if e.Key != "" {
			k2vToFetch = append(k2vToFetch, e.Key)
		} else {
			v2kToFetch = append(v2kToFetch, strconv.Itoa(e.Value))
		}
	}
	// do lookup on both
	_, errors := GetEntries(s.K2v, k2vToFetch)
	_, errorsTemp := GetEntries(s.V2k, v2kToFetch)
	// log errors
	for _, e := range errorsTemp {
		errors = append(errors, e)
	}
	for _, e := range errors {
		logErr(e.Error())
	}
	// combine into entries array

}

// create new entry in db
func (s *Server) CreateEntry(c *gin.Context) {

}

// export db to file
func (s *Server) ExportDB(c *gin.Context) {

}
