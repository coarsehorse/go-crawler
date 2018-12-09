package main

import (
	"encoding/json"
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
	"strings"
	"time"
)

const (
	LOG_FILENAME = "log.log"
)

func main() {
	// Input variations
	//url := "https://beteastsports.com/"
	//url := "https://www.sportintan.com/"
	//url := "https://ampmlimo.ca/"
	//url := "https://www.polygon.com/playstation"
	//url := "https://mediglobus.com/"
	//url := "http://example.com/"
	url := "https://www.study.ua/"
	//url := "https://www.study.ua/q/"
	//url := "https://www.study.ua/q/#"

	// Initialize logger
	/*logFilename, err := os.OpenFile(LOG_FILENAME, os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFilename)*/

	// Check time
	start := time.Now()

	// Read sitemap
	sitemap, err := crawler.GetLinksFromSitemap(url)
	utils.CheckError(err)
	linksToCrawl := utils.UniqueStringSlice(append(sitemap, url))
	// Crawl specified url
	crawledLevels := crawler.Crawl(linksToCrawl, []string{}, []crawler.CrawledLevel{})

	// Get execution time in ms
	executionTime := time.Now().Sub(start).Nanoseconds() / 1E+6

	// Marshal result
	marshaled, err := json.MarshalIndent(crawledLevels, "", "\t") // marshal to json with indents
	utils.CheckError(err)

	// Create unique backup file
	file, err := utils.CreateUniqResultingFile(url, ".json")
	utils.CheckError(err)

	// Write out result
	err = utils.WriteToFileAndClose(file, marshaled)
	utils.CheckError(err)

	// Create the file for crawled links only file
	crawledLinks := crawler.ExtractUniqueLinks(crawledLevels)
	f, err := utils.CreateUniqResultingFile(url, "-links-only.txt")
	utils.CheckError(err)
	err = utils.WriteToFileAndClose(f, []byte(strings.Join(crawledLinks, "\n")))
	utils.CheckError(err)

	log.Println("Execution time: ", executionTime, " ms")
}
