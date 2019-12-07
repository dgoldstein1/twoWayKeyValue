package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(t *testing.T) {

}

func TestParseEnv(t *testing.T) {
	// mock out log.Fatalf
	origLogFatalf := logFatalf
	defer func() { logFatalf = origLogFatalf }()
	errors := []string{}
	logFatalf = func(format string, args ...interface{}) {
		if len(args) > 0 {
			errors = append(errors, fmt.Sprintf(format, args))
		} else {
			errors = append(errors, format)
		}
	}

	originLogPrintf := logMsg
	defer func() { logMsg = originLogPrintf }()
	logs := []string{}
	logMsg = func(format string, args ...interface{}) {
		if len(args) > 0 {
			logs = append(logs, fmt.Sprintf(format, args))
		} else {
			logs = append(logs, format)
		}
	}

	requiredEnvs := []string{
		"GRAPH_DB_STORE_DIR",
		"GRAPH_DB_STORE_PORT",
		"GRAPH_DOCS_DIR",
	}

	for _, v := range requiredEnvs {
		os.Setenv(v, "1000")
	}
	// positive test
	parseEnv()
	assert.Equal(t, len(errors), 0)

	for _, v := range requiredEnvs {
		t.Run("it validates "+v, func(t *testing.T) {
			errors = []string{}
			os.Unsetenv(v)
			parseEnv()
			assert.Equal(t, len(errors) > 0, true)
			// cleanup
			os.Setenv(v, "5")
		})
	}

	t.Run("sets GRAPH_DB_STORE_PORT to PORT if is unset", func(t *testing.T) {
		errors = []string{}
		for _, v := range requiredEnvs {
			os.Setenv(v, "5")
		}
		os.Unsetenv("GRAPH_DB_STORE_PORT")
		os.Setenv("PORT", "15367")
		parseEnv()
		assert.Equal(t, errors, []string{})
		assert.Equal(t, os.Getenv("GRAPH_DB_STORE_PORT"), "15367")
	})

	t.Run("fails if GRAPH_DB_STORE_PORT is not valid int", func(t *testing.T) {
		errors = []string{}
		os.Setenv("GRAPH_DB_STORE_PORT", "f232")
		parseEnv()
		assert.Equal(t, 2, len(errors))
		assert.Equal(t, "strconv.Atoi: parsing \"f232\": invalid syntax", errors[0])
		assert.Equal(t, "GRAPH_DB_STORE_PORT must be a valid port in range but was '[%!i(int=0)]'", errors[1])
	})
	t.Run("fails if GRAPH_DB_STORE_PORT is not a positive int", func(t *testing.T) {
		errors = []string{}
		os.Setenv("GRAPH_DB_STORE_PORT", "-253")
		parseEnv()
		assert.Equal(t, 1, len(errors))
		assert.Equal(t, "GRAPH_DB_STORE_PORT must be a valid port in range but was '[%!i(int=-253)]'", errors[0])
	})
	t.Run("throws no errors if GRAPH_DB_STORE_PORT is '2534'", func(t *testing.T) {
		errors = []string{}
		os.Setenv("GRAPH_DB_STORE_PORT", "2534")
		parseEnv()
		assert.Equal(t, 0, len(errors))
	})
}
