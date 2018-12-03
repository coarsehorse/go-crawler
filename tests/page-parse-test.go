package main

import (
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
)

func main() {
	url := "https://www.study.ua/"
	page, err := crawler.ParsePage(url)
	log.Print(page)
	utils.CheckError(err)
}
