package server

import (
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
func ListenAndServe(port int) {
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
	router.LoadHTMLGlob("api/*.html")
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
	// start server
	logMsg("Serving on port %d", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		logFatalf("ListenAndServe: %v", err)
	}
}

// retrieve and try from db
func (s *Server) RetreieveEntry(c *gin.Context) {

}

// create new entry in db
func (s *Server) CreateEntry(c *gin.Context) {

}

// export db to file
func (s *Server) ExportDB(c *gin.Context) {

}
