package main

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"strconv"
	"time"
	//"strings"
)

func parsePage1(url string) (result string, err error) {
	resp, err := http.Get(url)

	// Handle response errors
	if err != nil {
		errMessage := "Failed to crawl1 " + url + " with error: \"" + err.Error() + "\""
		return "", errors.New(errMessage)
	}

	// Handle not 200 status
	if resp.StatusCode != 200 {
		errMessage := "Failed to crawl1 " + url + " with error: \"Not 200 status code(" + strconv.Itoa(resp.StatusCode) + ")\""
		return "", errors.New(errMessage)
	}

	respBody := resp.Body
	defer respBody.Close()

	// data
	var h1 string
	var title string
	var links = make([]string, 0)
	var imgs = make([]string, 0)
	//

	htmlNode, err := html.Parse(respBody)

	if err != nil {
		errMessage := "Failed to parse HTML node on " + url
		return "", errors.New(errMessage)
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		for c := n.NextSibling; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "h1" {
				h1 = c.FirstChild.Data
			}
			if c.Type == html.ElementNode && c.Data == "title" {
				h1 = c.FirstChild.Data
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(htmlNode)

	//errMessage := "No HTML tokens found on " + url
	//return "", errors.New(errMessage)

	fmt.Println(h1)
	fmt.Println(title)
	fmt.Println(links)
	fmt.Println(imgs)

	return "", nil
}

func main() {
	url := "https://beteastsports.com"
	title, err := parsePage(url)

	start := time.Now()

	fmt.Println(err, title)

	elapsed := time.Since(start)
	fmt.Printf("Binomial took %s s\n", elapsed)
}
