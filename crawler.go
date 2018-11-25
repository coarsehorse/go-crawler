package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	RESULTS_DIR   = "RESULTS"
	GORUTINES_NUM = 32
)

type CrawledPage struct {
	Url            string            `json:"url"`
	H1             string            `json:"h1"`
	Title          string            `json:"title"`
	Links          []string          `json:"links"`
	HreflangUrlMap map[string]string `json:"hreflangUrlMap"`
	Imgs           []string          `json:"imgs"`
	CanonicalUrl   string            `json:"canonicalUrl"`
	NoIndex        bool              `json:"noIndex"`
}

type CrawledLevel struct {
	LevelNum     int           `json:"levelNum"`
	CrawledPages []CrawledPage `json:"crawledPages"`
}

func parsePage(url string) (CrawledPage, error) {
	// Ensure url is ok
	url = addFollowingSlash(url)

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

	// Fill url
	crawledPage.Url = url

	// Grab title
	title := doc.Find("title").Eq(0).Text()
	title = strings.TrimSpace(title)
	crawledPage.Title = title

	// Grab h1
	h1 := doc.Find("h1").Eq(0).Text()
	h1 = strings.TrimSpace(h1)
	crawledPage.H1 = h1

	// Grab links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		link = strings.TrimSpace(link)
		if exists {
			crawledPage.Links = append(crawledPage.Links, link)
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

		crawledPage.HreflangUrlMap[hreflang] = href
	})

	// Grab imgs
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		imgSrc, exists := s.Attr("src")
		imgSrc = strings.TrimSpace(imgSrc)
		if exists {
			crawledPage.Imgs = append(crawledPage.Imgs, imgSrc)
		}
	})

	// Grab canonical url
	canonicalUrl, exists := doc.Find("link[rel *= 'canonical']").Eq(0).Attr("href")
	if exists {
		canonicalUrl = strings.TrimSpace(canonicalUrl)
		crawledPage.CanonicalUrl = canonicalUrl
	}

	// Grab noindex
	_, exists = doc.Find("meta[content *= 'noindex']").Eq(0).Attr("content")
	if exists {
		crawledPage.NoIndex = true
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
		lastLevelNum = crawledLevels[0].LevelNum
	} else {
		lastLevelNum = -1
	}
	crawledLevels = append(crawledLevels, CrawledLevel{
		LevelNum:     lastLevelNum + 1,
		CrawledPages: crawledPages,
	})

	// Unique crawledLinks
	crawledLinks = uniqueStringSlice(crawledLinks)

	// Collect and unique all links from crawled pages
	nextLevelLinksMap := make(map[string]struct{}, 0)
	for _, page := range crawledPages {
		for _, link := range page.Links {
			nextLevelLinksMap[link] = struct{}{}
		}
	}
	// Convert map to a slice
	nextLevelLinks := make([]string, 0, len(nextLevelLinksMap))
	for k := range nextLevelLinksMap {
		nextLevelLinks = append(nextLevelLinks, k)
	}
	// Filter out bad links(tel:, mailto:, #, etc.)
	nextLevelLinks = filterSlice(nextLevelLinks, func(link string) bool {
		if link == "" || link == "#" {
			return false
		} else if strings.HasPrefix(link, "tel:") ||
			strings.HasPrefix(link, "mailto:") ||
			strings.Contains(link, "javascript:void(0)") ||
			strings.Contains(link, "javascript:;") {

			return false
		}
		return true
	})

	// Add following "/"
	foo := make([]string, 0, len(nextLevelLinks))
	for _, link := range nextLevelLinks {
		foo = append(foo, addFollowingSlash(link))
	}
	nextLevelLinks = foo

	// Extend relative links
	foo = make([]string, 0, len(nextLevelLinks))
	for _, link := range nextLevelLinks {
		if extendedLink, err := extendRelativeLink(link, linksToCrawl[0]); err == nil {
			foo = append(foo, extendedLink)
		}
	}
	nextLevelLinks = foo

	// Remove duplicates
	nextLevelLinks = uniqueStringSlice(nextLevelLinks)

	// Validate with domain pattern
	domain := extractDomain(linksToCrawl[0])
	domainPattern := `^(http|https):\/\/` + strings.Replace(domain, `.`, `\.`, -1) + `.*$`
	r, err := regexp.Compile(domainPattern)
	if err != nil {
		panic("Bad regexp constructed for links validation: \"" + domainPattern + "\"")
	}
	nextLevelLinks = filterSlice(nextLevelLinks, func(link string) bool {
		return r.MatchString(link) // check domain domainPattern
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

	if len(remainingLinks) == 0 { // crawling is done
		return crawledLevels
	} else {
		return crawl(remainingLinks, crawledLinks, crawledLevels) // crawl next level
	}
}

func uniqueStringSlice(initialSlice []string) []string {
	fooMap := make(map[string]struct{}, len(initialSlice))
	for _, str := range initialSlice {
		fooMap[str] = struct{}{}
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

func extractDomain(url string) string {
	return strings.Split(url, "/")[2]
}

// Creates the new result file.
// Filename mask: curDir + RESULTS_DIR + domain + date + ext.
// Where:
// curDir - current dir of execution,
// RESULTS_DIR - constant, represents name of directory with results
// domain - extracted url domain
// date - current date in format dd-MM-YYYY-HH-mm-ss
// ext - future file extension, leading dot is needed(.json, .jpg)
// Returns the *File pointer on the created file or nil on error + optional error
func createUniqResultFile(url string, ext string) (createdFile *os.File, err error) {
	domain := extractDomain(url)
	t := time.Now()
	date := t.Format("2-1-2006-15-04-05") // get datetime in string(dd-MM-YYYY-HH-mm-ss)
	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	_ = os.MkdirAll(filepath.Join(curDir, RESULTS_DIR), os.ModePerm)    // create results dir
	fileName := filepath.Join(curDir, RESULTS_DIR, domain+"-"+date+ext) // create res file absolute path
	createdFile, err = os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return createdFile, nil
}

// Writes data to the already created file with error handling
// and closing opened file after writing
// Returns the possible error or nil on success
func writeToFileAndClose(file *os.File, data []byte) error {
	_, err := file.Write(data) // write out result
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func addFollowingSlash(str string) string {
	if chars := strings.Split(str, ""); chars[len(chars)-1] != "/" {
		return str + "/"
	}
	return str
}

func extendRelativeLink(relativeLink string, linkAbsoluteLocation string) (absoluteUrl string, err error) {
	// If relativeLink is empty
	if relativeLink == "" {
		return "", errors.New("Empty relativeLink")
	}

	// To bee sure that input URLs has trailing '/'
	relativeLink = addFollowingSlash(relativeLink)
	linkAbsoluteLocation = addFollowingSlash(linkAbsoluteLocation)

	// Common data for the all cases
	absoluteSplitted := strings.Split(linkAbsoluteLocation, `/`)
	protocol := absoluteSplitted[0] + "//"
	domain := addFollowingSlash(extractDomain(linkAbsoluteLocation))

	// Case /
	if relativeLink == `/` { // root
		return protocol + domain, nil
	}

	// Case relativeLink is already absolute
	pattern := `^(http|https):\/\/.*$`
	r := regexp.MustCompile(pattern)
	if r.MatchString(relativeLink) {
		return relativeLink, nil
	}

	// Case a/, path/to/page/, about/
	r = regexp.MustCompile(`^([\w-]+/)+$`)
	if r.MatchString(relativeLink) { // relative to location
		return linkAbsoluteLocation + relativeLink, nil
	}

	// Case /a/, /path/to/page/, /about/
	r = regexp.MustCompile(`^/(([\w-]+/)+)$`)
	if newRelative := r.FindStringSubmatch(relativeLink); newRelative != nil { // relative to root
		return protocol + domain + newRelative[1], nil
	}

	// Case //path/to/smth
	r = regexp.MustCompile(`^//(.*/)$`)
	if newLink := r.FindStringSubmatch(relativeLink); newLink != nil { // same protocol link
		return protocol + newLink[1], nil
	}

	// Case ../path/
	r = regexp.MustCompile(`^\.\./(.*/)$`)
	if newRelative := r.FindStringSubmatch(relativeLink); newRelative != nil {
		newRelativePart := newRelative[1]
		if len(absoluteSplitted) < 4 { // less than 4 levels in absolute: (http:)()(domain.com)(first_level)
			return "", errors.New("Cannot resolve '../' relative path for absolute location " +
				linkAbsoluteLocation)
		} else {
			newAbsoluteSplitted := absoluteSplitted[:len(absoluteSplitted)-2] // -2 because we have following /
			newAbsolute := addFollowingSlash(strings.Join(newAbsoluteSplitted, "/"))
			return newAbsolute + newRelativePart, nil
		}
	}

	return "", errors.New("Can't parse relative link: \"" + relativeLink + "\"")
}

func main() {
	//url := "https://beteastsports.com/"
	url := "https://ampmlimo.ca/"
	//url := "https://www.polygon.com/playstation"
	//url := "https://mediglobus.com/"
	//url := "http://example.com/"

	// Create unique backup file for future result
	file, err := createUniqResultFile(url, ".json")
	if err != nil {
		panic(err.Error())
	}

	start := time.Now() // check time

	//crawledPage, _ := parsePage(url)
	crawledLevels := crawl([]string{url}, []string{}, []CrawledLevel{})

	executionTime := time.Now().Sub(start).Nanoseconds() / 1E+6 // get execution time in ms

	//fmt.Printf("Crawled levels: %v\n", crawledPage)
	//fmt.Printf("Crawled levels: %v\n", crawledLevels)

	marshaled, err := json.MarshalIndent(crawledLevels, "", "\t") // marshal to json with indents
	if err != nil {
		panic(err.Error())
	}

	err = writeToFileAndClose(file, marshaled) // write out result
	if err != nil {
		panic(err.Error())
	}

	// Check res
	crawledLinks := make([]string, 0)
	for _, lvl := range crawledLevels {
		for _, page := range lvl.CrawledPages {
			crawledLinks = append(crawledLinks, page.Url)
		}
	}
	f, e := createUniqResultFile(url, "-GO-.txt")
	if e != nil {
		panic(e.Error())
	}
	e = writeToFileAndClose(f, []byte(strings.Join(crawledLinks, "\n")))
	if e != nil {
		panic(e.Error())
	}

	fmt.Printf("Execution time: %v ms\n", executionTime)
}
