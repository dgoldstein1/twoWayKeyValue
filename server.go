package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
)

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
	router.POST("/entries", s.CreateEntries)
	router.GET("/export", s.ExportDB)
	// return server
	return router, &s
}

// removes duplicate keys in array
func removeDuplicates(keys []string) (noDuplicates []string) {
	noDuplicates = []string{}
	m := make(map[string]bool)
	for _, k := range keys {
		if !m[k] {
			noDuplicates = append(noDuplicates, k)
			m[k] = true
		}
	}
	return noDuplicates
}

// create entries if they don't already exist
func (s *Server) CreateEntries(c *gin.Context) {
	// read in request
	keysToCreate := []string{}
	if err := c.BindJSON(&keysToCreate); err != nil {
		c.JSON(400, Error{400, err.Error()})
		return
	}
	if len(keysToCreate) == 0 {
		c.JSON(400, "Bad []entry or no entries passed")
		return
	}
	// remove duplicates from keys passed
	keysToCreate = removeDuplicates(keysToCreate)
	// finally return everything!!
	c.JSON(200, RetrieveEntryResponse{})
}

// stream zipped file over browser
func (s *Server) ExportDB(c *gin.Context) {
	fileName, err := ZipDb()
	if err != nil {
		logErr("Could not create zip %v", err)
		c.JSON(500, Error{500, err.Error()})
		return
	}
	c.FileAttachment(fileName, "twowaykv_export.zip")
}
