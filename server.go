package main

import (
	"github.com/gin-gonic/gin"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
	"strconv"
)

// entrypoint
func SetupRouter(docs string) (*gin.Engine, *Server) {
	// try to connect to db
	logMsg("Connecting to DB")
	kDB, vDB, err := ConnectToDb()
	logMsg("Done.")
	if err != nil {
		logFatalf("Could not establish connection to db: %v", err)
	}
	// create server object
	s := Server{kDB, vDB, CreateIfDoesntExist}
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
	router.POST("/entriesFromKeys", s.GetEntriesFromKeys)
	router.POST("/entriesFromValues", s.GetEntriesFromValues)
	router.GET("/random", s.RandomEntries)
	router.GET("/search", s.Search)
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
	// create dbs
	entries, errors := CreateIfDoesntExist(
		removeDuplicates(keysToCreate),              // remove duplicates from keys passed
		c.Query("muteAlreadyExistsError") == "true", // log or dont log already exists errors
		s.K2v,
		s.V2k,
	)
	// finally return everything!!
	c.JSON(200, RetrieveEntryResponse{errors, entries})
}

// Get a specified number of random entries
var MAX_N = 25

func (s *Server) RandomEntries(c *gin.Context) {
	n, err := strconv.Atoi(c.DefaultQuery("n", "1"))
	if err != nil {
		c.JSON(400, Error{400, "Invalid int"})
		return
	}
	if n > 25 || n < 1 {
		c.JSON(400, Error{400, "'n' must be positive and greater than " + strconv.Itoa(MAX_N)})
		return
	}
	entries, err := readRandomEntries(s.V2k, n)
	if err != nil {
		c.JSON(500, Error{500, err.Error()})
		return
	}
	// success
	c.JSON(200, entries)
}

func (s *Server) GetEntriesFromKeys(c *gin.Context) {
	keys := []string{}
	if err := c.BindJSON(&keys); err != nil {
		c.JSON(400, Error{400, err.Error()})
		return
	}
	keys = removeDuplicates(keys)
	entries, errs := GetEntriesFromKeys(s.K2v, keys)
	c.JSON(200, RetrieveEntryResponse{errs, entries})
}

func (s *Server) GetEntriesFromValues(c *gin.Context) {
	values := []int{}
	if err := c.BindJSON(&values); err != nil {
		c.JSON(400, Error{400, err.Error()})
		return
	}
	entries, errs := GetEntriesFromValues(s.V2k, values)
	c.JSON(200, RetrieveEntryResponse{errs, entries})
}

func (s *Server) Search(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(400, Error{400, "a query must be passed to /search"})
		return
	}
	entries, errs := SeekWithPrefix(s.K2v, q)
	c.JSON(200, RetrieveEntryResponse{errs, entries})
}
