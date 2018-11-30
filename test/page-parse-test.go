package main

import (
	"goCrawler/crawler"
	"goCrawler/utils"
	"log"
)

func main() {
	url := "https://beteastsports.com"
	page, err := crawler.ParsePage(url)
	log.Print(page)
	utils.CheckError(err)
}
