package server

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
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
	router.PUT("/entry", s.CreateEntry)
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
	fmt.Println(c.Request.Body)
	// read in request
	entries := []Entry{}
	if err := c.BindJSON(&entries); err != nil {
		c.JSON(400, Error{400, err.Error()})
	}
	if len(entries) == 0 {
		c.JSON(400, "Bad []entry or no entries passed")
	}
}

// create new entry in db
func (s *Server) CreateEntry(c *gin.Context) {

}

// export db to file
func (s *Server) ExportDB(c *gin.Context) {

}
