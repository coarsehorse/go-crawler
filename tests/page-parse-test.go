package main

import (
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
)

func main() {
	//url := "https://www.study.ua/"
	//url := "http://summer.study.ua/trip/listview/"
	//url := "https://www.study.ua/q/"
	//url := "https://www.study.ua/q/#"
	//url := "https://www.study.ua/program-7720.htm"
	url := "https://www.study.ua/program-275.htm"

	page, err := crawler.ParsePage(url)
	log.Print(page)
	utils.CheckError(err)
}
