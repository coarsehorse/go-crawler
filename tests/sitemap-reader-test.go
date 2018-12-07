package main

import (
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
)

func main() {
	url := "https://www.study.ua/"

	links, err := crawler.GetLinksFromSitemap(url)
	utils.CheckError(err)
	log.Print(links)
}
