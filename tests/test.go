package main

import (
	"go-crawler/crawler"
	"go-crawler/utils"
	"log"
)

func main() {
	//rel := "/q/#Embassy English — английский для детей в Канаде/"
	//rel := "/consultation/?theme=Артур%20Нещадин,%20отзыв/"
	//rel := "/q/#Albert-Ludwigs-Universität Freiburg/"
	//rel := "tel: 0970000320"
	rel := "news/start-edu-2019.htm"

	//abs := "https://www.study.ua/program-7819"
	//abs := "https://www.study.ua/program-7819.htm"
	abs := "https://www.study.ua/e/?f=Au-Pair-USA/"

	res, err := crawler.ExtendRelativeLink(rel, abs)
	utils.CheckError(err)
	log.Println(res)
}
