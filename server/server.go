package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
)

var logFatalf = log.Fatalf
var logMsg = log.Infof

// entrypoint
func ListenAndServe(port int) {
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
	router.GET("/entry", RetreieveEntry)
	router.POST("/entry", CreateEntry)
	router.GET("/save", ExportDB)
	// start server
	logMsg("Serving on port %d", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), router)
	if err != nil {
		logFatalf("ListenAndServe: %v", err)
	}
}

// retrieve and try from db
func RetreieveEntry(c *gin.Context) {

}

// create new entry in db
func CreateEntry(c *gin.Context) {

}

// export db to file
func ExportDB(c *gin.Context) {

}
