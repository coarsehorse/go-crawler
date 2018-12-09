package main

import (
	"fmt"
	"go-crawler/crawler"
)

func main() {
	// Test conversion

	// Initial data
	absLocat := "https://domain.com/one/two/three"
	fmt.Println("Initial data\n")
	fmt.Println("AbsoluteLocation: " + absLocat + "\n")

	// path/path1 case
	rel := "path/path1"
	res, err := crawler.ExtendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// ../path case
	rel = "../path"
	res, err = crawler.ExtendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("../path case")
	fmt.Println(res)
	fmt.Println()

	// /path/path1 case
	rel = "/path/path1"
	res, err = crawler.ExtendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("/path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// //path/path1 case
	rel = "//path/path1"
	res, err = crawler.ExtendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("//path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// Already absolute like http://domain.com/path case
	rel = "http://domain.com/path"
	res, err = crawler.ExtendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("http://domain.com/path case")
	fmt.Println(res)
	fmt.Println()
}
