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
		"GRAPH_DB_ENDPOINT",
		"STARTING_ENDPOINT",
		"MAX_APPROX_NODES",
	}

	for _, v := range requiredEnvs {
		os.Setenv(v, "5")
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

	t.Run("fails if MAX_APPROX_NODES is not valid int", func(t *testing.T) {
		errors = []string{}
		os.Setenv("MAX_APPROX_NODES", "f232")
		parseEnv()
		assert.Equal(t, 2, len(errors))
		assert.Equal(t, "strconv.Atoi: parsing \"f232\": invalid syntax", errors[0])
		assert.Equal(t, "MAX_APPROX_NODES must be greater than 1 but was '[%!i(int=0)]'", errors[1])
	})
	t.Run("fails if MAX_APPROX_NODES is not a positive int", func(t *testing.T) {
		errors = []string{}
		os.Setenv("MAX_APPROX_NODES", "-253")
		parseEnv()
		assert.Equal(t, 1, len(errors))
		assert.Equal(t, "MAX_APPROX_NODES must be greater than 1 but was '[%!i(int=-253)]'", errors[0])
	})
	t.Run("throws no errors if MAX_APPROX_NODES is '-1'", func(t *testing.T) {
		errors = []string{}
		os.Setenv("MAX_APPROX_NODES", "-1")
		parseEnv()
		assert.Equal(t, 0, len(errors))
	})
}
