package main

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CrawledPage struct {
	url            string
	h1             string
	title          string
	links          []string
	hreflangUrlMap map[string]string
	imgs           []string
	canonicalUrl   string
	noIndex        bool
}

type CrawledLevel struct {
	levelNum     int
	crawledPages []CrawledPage
}

func parsePage(url string) (CrawledPage, error) {
	// Get page by url
	//start := time.Now()
	resp, err := http.Get(url)
	//executionTime := time.Now().Sub(start).Nanoseconds() / 10E+6 // ms
	//fmt.Printf("Downloaded by: %v ms\n", executionTime)

	// Handle response errors
	if err != nil {
		errMessage := "Failed to crawl1 " + url + " with error: \"" + err.Error() + "\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Handle not 200 status
	if resp.StatusCode != 200 {
		errMessage := "Failed to crawl1 " + url + " with error: \"Not 200 status code(" + strconv.Itoa(resp.StatusCode) + ")\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Create goquery Document
	respBodyReader := resp.Body
	defer respBodyReader.Close()

	doc, err := goquery.NewDocumentFromReader(respBodyReader)

	if err != nil {
		errMessage := "Failed to create goquery Document from " + url + " with error: \"" + err.Error() + "\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Init future result
	crawledPage := CrawledPage{"", "", "", make([]string, 0),
		make(map[string]string), make([]string, 0), "", false}

	/* Find data */

	// Grab title
	title := doc.Find("title").Eq(0).Text()
	title = strings.TrimSpace(title)
	crawledPage.title = title

	// Grab h1
	h1 := doc.Find("h1").Eq(0).Text()
	h1 = strings.TrimSpace(h1)
	crawledPage.h1 = h1

	// Grab links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		link = strings.TrimSpace(link)
		if exists {
			crawledPage.links = append(crawledPage.links, link)
		}
	})

	// Grab hreflangs
	doc.Find("link[rel *= 'alternate']").Each(func(i int, s *goquery.Selection) {
		hreflang, exists := s.Attr("hreflang")
		if !exists {
			return
		}

		href, exists := s.Attr("href")
		if !exists {
			return
		}

		hreflang = strings.TrimSpace(hreflang)
		href = strings.TrimSpace(href)

		crawledPage.hreflangUrlMap[hreflang] = href
	})

	// Grab imgs
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		imgSrc, exists := s.Attr("src")
		imgSrc = strings.TrimSpace(imgSrc)
		if exists {
			crawledPage.imgs = append(crawledPage.imgs, imgSrc)
		}
	})

	// Grab canonical url
	canonicalUrl, exists := doc.Find("link[rel *= 'canonical']").Eq(0).Attr("href")
	if exists {
		canonicalUrl = strings.TrimSpace(canonicalUrl)
		crawledPage.canonicalUrl = canonicalUrl
	}

	// Grab noindex
	_, exists = doc.Find("meta[content *= 'noindex']").Eq(0).Attr("content")
	if exists {
		crawledPage.noIndex = true
	}

	return crawledPage, nil
}

func crawl(linksToCrawl []string, crawledLinks []string, crawledLevels []CrawledLevel) []CrawledLevel {
	notGotPages := 0
	crawledPages := make([]CrawledPage, 0)
	for _, link := range linksToCrawl {
		page, err := parsePage(link)
		if err != nil {
			notGotPages++
		} else {
			crawledPages = append(crawledPages, page)
		}
		crawledLinks = append(crawledLinks, link)
	}
	fmt.Printf("Expected to be crawled: %v, not crawled: %v\n", len(linksToCrawl), notGotPages)

	// Add the new crawled level to crawledLevels
	var lastLevelNum int
	if len(crawledLevels) > 0 {
		lastLevelNum = crawledLevels[0].levelNum
	} else {
		lastLevelNum = -1
	}
	crawledLevels = append(crawledLevels, CrawledLevel{
		levelNum:     lastLevelNum + 1,
		crawledPages: crawledPages,
	})

	// Unique crawledLinks
	crawledLinks = uniqueStringSlice(crawledLinks)

	// Collect and unique all links from crawled pages
	nextLevelLinksMap := make(map[string]bool, 0)
	for _, page := range crawledPages {
		for _, link := range page.links {
			nextLevelLinksMap[link] = true
		}
	}
	// Convert map to a slice
	nextLevelLinks := make([]string, 0, len(nextLevelLinksMap))
	for k := range nextLevelLinksMap {
		nextLevelLinks = append(nextLevelLinks, k)
	}
	// Filter out links from another domains
	domain := strings.Split(linksToCrawl[0], "/")[2]
	pattern := `^(http|https):\/\/` + strings.Replace(domain, `.`, `\.`, -1) + `.*$`
	r, err := regexp.Compile(pattern)

	if err != nil {
		panic("Bad regexp constructed for links validation: \"" + pattern + "\"")
	}
	nextLevelLinks = filterSlice(nextLevelLinks, func(link string) bool {
		return r.MatchString(link)
	})

	// Convert crawledLinks to map to be able to search in it
	crawledMap := make(map[string]struct{}, len(crawledLinks))
	for _, link := range crawledLinks {
		crawledMap[link] = struct{}{}
	}
	// Separate not crawled links from nextLevelLinks
	remainingLinks := make([]string, 0, len(nextLevelLinks))
	for _, link := range nextLevelLinks {
		_, ok := crawledMap[link]
		if !ok { // link is not already crawled
			remainingLinks = append(remainingLinks, link)
		}
	}

	if len(remainingLinks) == 0 {
		return crawledLevels
	} else {
		return crawl(remainingLinks, crawledLinks, crawledLevels)
	}
}

func uniqueStringSlice(initialSlice []string) []string {
	fooMap := make(map[string]bool, len(initialSlice))
	for _, str := range initialSlice {
		fooMap[str] = true
	}
	uniqueSlice := make([]string, 0, len(fooMap))
	for k := range fooMap {
		uniqueSlice = append(uniqueSlice, k)
	}

	return uniqueSlice
}

func filterSlice(slice []string, predicate func(string) bool) (filtered []string) {
	for _, s := range slice {
		if predicate(s) {
			filtered = append(filtered, s)
		}
	}
	return
}

func main() {
	//url := "https://beteastsports.com/"
	url := "https://www.polygon.com/playstation"
	//url := "https://mediglobus.com/"

	start := time.Now()

	//crawledPage, _ := parsePage(url)
	crawledLevels := crawl([]string{url}, []string{}, []CrawledLevel{})
	//fmt.Printf("Crawled levels: %v\n", crawledPage)
	fmt.Printf("Crawled levels: %v\n", crawledLevels)

	executionTime := time.Now().Sub(start).Nanoseconds() / 10E+6 // ms

	fmt.Printf("Execution time: %v ms\n", executionTime)
}
