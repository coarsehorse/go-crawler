package main

import (
	"go-crawler/utils"
	"log"
	"regexp"
)

func main() {
	url1 := "https://study.ua/"
	url2 := "http://summer.winter.study.ua/"
	url3 := "http://nestudy.ua/asd-study.ua/"

	pattern := `^https?:\/\/([^\.]+\.)+study\.ua\/.*$`
	r, err := regexp.Compile(pattern)
	utils.CheckError(err)

	log.Print(r.MatchString(url1))
	log.Print(r.MatchString(url2))
	log.Print(r.MatchString(url3))

	a := []string{"asd", "asd1", "asd2"}
	log.Print(a[len(a)-2 : len(a)])
}
