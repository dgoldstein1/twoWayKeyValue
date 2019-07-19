package server

import (
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
func SetupRouter() (*gin.Engine, *Server) {
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
	if gin.Mode() != gin.TestMode {
		router.LoadHTMLGlob("api/*.html")
		router.Static("/static", "static")
		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
	}
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

// retrieve and try from db
func (s *Server) RetreieveEntry(c *gin.Context) {
	// validate that either key of value was passed
	key := c.Query("key")
	val := c.Query("value")
	if key == "" && val == "" {
		c.JSON(400, Error{
			Error: "Must provide either a key or value query string",
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
