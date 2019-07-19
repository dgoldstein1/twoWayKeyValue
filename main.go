package main

import (
	"github.com/dgoldstein1/crawler/crawler"
	wiki "github.com/dgoldstein1/crawler/wikipedia"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

// checks environment for required env vars
var logFatalf = log.Fatalf
var logMsg = log.Infof

func parseEnv() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	requiredEnvs := []string{
		"GRAPH_DB_ENDPOINT",
		"STARTING_ENDPOINT",
		"MAX_APPROX_NODES",
	}
	for _, v := range requiredEnvs {
		if os.Getenv(v) == "" {
			logFatalf("'%s' was not set", v)
		} else {
			// print out config
			logMsg("%s=%s", v, os.Getenv(v))
		}
	}
	i, err := strconv.Atoi(os.Getenv("MAX_APPROX_NODES"))
	if err != nil {
		logFatalf(err.Error())
	}
	if i < 1 && i != -1 {
		logFatalf("MAX_APPROX_NODES must be greater than 1 but was '%i'", i)
	}
}

// runs crawler with given functions
func runCrawler(
	isValidCrawlLink crawler.IsValidCrawlLinkFunction,
	connectToDB crawler.ConnectToDBFunction,
	addEdgeIfDoesNotExist crawler.AddEdgeFunction,
) {
	// assert environment
	parseEnv()
	// crawl with passed args
	MAX_APPROX_NODES, _ := strconv.Atoi(os.Getenv("MAX_APPROX_NODES"))
	crawler.Crawl(
		os.Getenv("STARTING_ENDPOINT"),
		int32(MAX_APPROX_NODES),
		isValidCrawlLink,
		connectToDB,
		addEdgeIfDoesNotExist,
	)
}

func main() {
	app := cli.NewApp()
	app.Name = "crawler"
	app.Usage = " acustomizable web crawler script for different websites"
	app.Description = "web crawl different URLs and add similar urls to a graph database"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:    "wikipedia",
			Aliases: []string{"w"},
			Usage:   "crawl on wikipedia articles",
			Action: func(c *cli.Context) error {
				runCrawler(
					wiki.IsValidCrawlLink,
					wiki.ConnectToDB,
					wiki.AddEdgesIfDoNotExist,
				)
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
