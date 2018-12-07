package main

import (
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
)

func main() {
	//url := "https://www.study.ua/"
	url := "http://summer.study.ua/trip/listview/"

	page, err := crawler.ParsePage(url)
	log.Print(page)
	utils.CheckError(err)
}
