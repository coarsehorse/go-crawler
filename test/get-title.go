package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
)

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func isH1Element(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "h1"
}

func traverse(n *html.Node) (string, bool) {
	//if isTitleElement(n) {
	//	return n.FirstChild.Data, true
	//}
	if isH1Element(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func GetHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		panic("Fail to parse html")
	}

	return traverse(doc)
}

func main() {
	url := "https://beteastsports.com"
	resp, _ := http.Get(url)
	str, _ := GetHtmlTitle(resp.Body)
	fmt.Println(str)
}
