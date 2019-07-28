package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/zsais/go-gin-prometheus"
	"net/http"
	"strconv"
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
	router.POST("/entries", s.PostEntries)
	router.GET("/export", s.ExportDB)
	// return server
	return router, &s
}

// validate key and value
func ValidatEntry(e Entry) error {
	if e.Key == "" && e.Value == 0 {
		return errors.New("Must provide valid key or value query string")
	}
	if e.Value < 0 {
		return fmt.Errorf("Invalid int '%d'", e.Value)
	}
	return nil
}

// retrieve and try from db
func (s *Server) PostEntries(c *gin.Context) {
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
	// lookup all
	entriesToReturn := []Entry{}
	k2vEntries, k2vErrors := GetEntries(s.K2v, k2vToFetch)
	v2kEntries, v2kErrors := GetEntries(s.V2k, v2kToFetch)
	// pool errors
	errors := []string{}
	for _, e := range v2kErrors {
		errors = append(errors, e.Error)
		logErr(e.Error)
	}
	// add to db if not found
	for _, e := range k2vErrors {
		if !e.NotFound {
			errors = append(errors, e.Error)
		} else {
			// create new, key not found
			entry, err := s.WriteEntry(s.K2v, s.V2k, e.LookupId)
			if err != nil {
				errors = append(errors, err.Error())
				logErr(err.Error())
			} else {
				entriesToReturn = append(entriesToReturn, entry)
			}
		}
	}
	// combine into entries array
	for key, v := range k2vEntries {
		val, err := strconv.Atoi(v)
		if err != nil {
			errors = append(errors, err.Error())
			logErr("Could not convert value to int %s", v)
		} else {
			entriesToReturn = append(entriesToReturn, Entry{key, val})
		}
	}
	for v, key := range v2kEntries {
		val, _ := strconv.Atoi(v)
		entriesToReturn = append(entriesToReturn, Entry{key, val})
	}
	// finally return everything!!
	c.JSON(200, RetrieveEntryResponse{errors, entriesToReturn})
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
