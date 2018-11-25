package main

import (
	"fmt"
	"github.com/opesun/goquery"
	"strings"
	"time"
)

var (
	WORKERS int = 4 // goroutines quantity
)

func parse_page(url string) (result string, err error) {
	x, err := goquery.ParseUrl(url)

	if err == nil {
		if s := strings.TrimSpace(x.Find(".fi_text").Text()); s != "" {
			return s, nil
		}
	}

	return "", err
}

func grab() <-chan string {
	c := make(chan string)
	for i := 0; i < WORKERS; i++ {
		go func() {
			for {
				time.Sleep(1000 * time.Millisecond)
				c <- "String from another thread"
			}
		}()
	}
	fmt.Println("Goroutines runned: ", WORKERS)
	return c
}

func main() {
	/*quotes_chan := grab()

	for {
		select {
		case str := <- quotes_chan:
			fmt.Println(str)
		}
	}*/

	/*url := "http://vpustotu.ru/moderation/"
	str, err := parse_page(url)

	if err == nil {
		fmt.Println(str);
	}*/

	/*s := `<p>Links:</p><ul><li><a href="foo">Foo</a><li><a href="/bar/baz">BarBaz</a></ul>`
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		log.Fatal(err)
	}

	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {

			for _, a := range n.Attr {
				if a.Key == "href" {
					fmt.Println(a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			//f(c)
			fmt.Println(c)
		}
	}

	f(doc)*/

	/*s := `<p>Links:</p><ul><li><a href="foo">Foo</a><li><a href="/bar/baz">BarBaz</a></ul>`

	tokenizer := html.NewTokenizer(strings.NewReader(s))

	for tokenType := tokenizer.Next(); tokenType != html.ErrorToken; tokenType = tokenizer.Next() {
		if tokenType != html.TextToken {
			continue
		}
		txtContent := strings.TrimSpace(html.UnescapeString(string(tokenizer.Text())))
		fmt.Println("# " + txtContent + " #" )
	}*/
}
