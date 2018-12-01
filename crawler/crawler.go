package crawler

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"go-crawler/utils"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
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

func ParsePage(url string) (CrawledPage, error) {
	// Check the time
	start := time.Now()

	// Ensure url is ok
	url = utils.AddFollowingSlash(url)

	// Get page by url
	resp, err := http.Get(url)

	// Handle response errors
	if err != nil {
		notifyAboutUrlWithTime(url, start, true, "")
		errMessage := "Failed to crawl1 " + url + " with error: \"" + err.Error() + "\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Handle not 200 status of original query or last redirect
	if resp.StatusCode != 200 {
		notifyAboutUrlWithTime(url, start, false, resp.Status)
		errMessage := "Failed to crawl1 " + url + " with error: \"Not 200 status code(" + strconv.Itoa(resp.StatusCode) + ")\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Create goquery Document
	respBodyReader := resp.Body
	doc, err := goquery.NewDocumentFromReader(respBodyReader)
	if err != nil {
		errMessage := "Failed to create goquery Document from " + url + " with error: \"" + err.Error() + "\""
		return CrawledPage{}, errors.New(errMessage)
	}

	// Init future result
	crawledPage := CrawledPage{"", "", "", make([]string, 0),
		make(map[string]string), make([]string, 0), "", false}

	/* Find data */

	// Grab url
	// Get original url or the last redirect
	crawledPage.Url = utils.AddFollowingSlash(strings.TrimSpace(resp.Request.URL.String()))

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
	// Extend relative links
	foo := make([]string, 0, len(crawledPage.Links))
	for _, link := range crawledPage.Links {
		if extendedLink, err := extendRelativeLink(link, url); err == nil {
			foo = append(foo, extendedLink)
		}
	}
	crawledPage.Links = foo

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
		href, err = extendRelativeLink(href, url)
		if err != nil {
			panic(err.Error())
		}

		crawledPage.HreflangUrlMap[hreflang] = href
		crawledPage.Links = append(crawledPage.Links, href)
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
		canonicalUrl, err = extendRelativeLink(strings.TrimSpace(canonicalUrl), url)
		if err != nil {
			return CrawledPage{}, err
		}
		crawledPage.CanonicalUrl = canonicalUrl
		crawledPage.Links = append(crawledPage.Links, canonicalUrl)
	}

	// Grab noindex
	_, exists = doc.Find("meta[content *= 'noindex']").Eq(0).Attr("content")
	if exists {
		crawledPage.NoIndex = true
	}

	// Checking pagination pattern
	r := regexp.MustCompile(`^((http|https):\/\/.*\/)(page|p)\/\d+\/$`)
	if paginationRootMatched := r.FindStringSubmatch(url); paginationRootMatched != nil {
		paginationRoot := utils.AddFollowingSlash(paginationRootMatched[1])
		crawledPage.Links = append(crawledPage.Links, paginationRoot)
	}

	// Checking get parameters pattern
	if strings.Contains(url, `?`) {
		withoutGet := utils.AddFollowingSlash(strings.Split(url, `?`)[0])
		crawledPage.Links = append(crawledPage.Links, withoutGet)
	}

	notifyAboutUrlWithTime(url, start, false, resp.Status)

	// Cleanup
	err = respBodyReader.Close()
	if err != nil {
		return CrawledPage{}, err
	}

	return crawledPage, nil
}

func Crawl(linksToCrawl []string, crawledLinks []string, crawledLevels []CrawledLevel) []CrawledLevel {
	log.Print("Starting crawl ", len(linksToCrawl), " links")
	notGotPages := 0
	crawledPages := make([]CrawledPage, 0)
	for _, link := range linksToCrawl {
		page, err := ParsePage(link)
		if err != nil {
			notGotPages++
		} else {
			crawledPages = append(crawledPages, page)
		}
		crawledLinks = append(crawledLinks, link)
		if link != page.Url { // if request was redirected
			crawledLinks = append(crawledLinks, page.Url)
		}
	}
	log.Print("Crawled with error ", notGotPages, "/", len(linksToCrawl), " links")

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
	crawledLinks = utils.UniqueStringSlice(crawledLinks)

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
	nextLevelLinks = utils.FilterSlice(nextLevelLinks, func(link string) bool {
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
		foo = append(foo, utils.AddFollowingSlash(link))
	}
	nextLevelLinks = foo

	// Remove duplicates
	nextLevelLinks = utils.UniqueStringSlice(nextLevelLinks)

	// Validate with domain pattern
	domain := utils.ExtractDomain(linksToCrawl[0])
	domainPattern := `^(http|https):\/\/` + strings.Replace(domain, `.`, `\.`, -1) + `.*$`
	r, err := regexp.Compile(domainPattern)
	if err != nil {
		panic("Bad regexp constructed for links validation: \"" + domainPattern + "\"")
	}
	nextLevelLinks = utils.FilterSlice(nextLevelLinks, func(link string) bool {
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
		return Crawl(remainingLinks, crawledLinks, crawledLevels) // crawl next level
	}
}

func extendRelativeLink(relativeLink string, linkAbsoluteLocation string) (absoluteUrl string, err error) {
	// If relativeLink is empty
	if relativeLink == "" {
		return "", errors.New("Empty relativeLink")
	}

	// To bee sure that input URLs has trailing '/'
	relativeLink = utils.AddFollowingSlash(relativeLink)
	linkAbsoluteLocation = utils.AddFollowingSlash(linkAbsoluteLocation)

	// Common data for the all cases
	absoluteSplitted := strings.Split(linkAbsoluteLocation, `/`)
	protocol := absoluteSplitted[0] + "//"
	domain := utils.AddFollowingSlash(utils.ExtractDomain(linkAbsoluteLocation))

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
			newAbsolute := utils.AddFollowingSlash(strings.Join(newAbsoluteSplitted, "/"))
			return newAbsolute + newRelativePart, nil
		}
	}

	return "", errors.New("Can't parse relative link: \"" + relativeLink + "\"")
}

func ExtractUniqueLinks(levels []CrawledLevel) (uniqueLinks []string) {
	uniqueLinksMap := make(map[string]struct{})

	for _, lvl := range levels {
		for _, page := range lvl.CrawledPages {
			uniqueLinksMap[page.Url] = struct{}{}
		}
	}
	for k, _ := range uniqueLinksMap {
		uniqueLinks = append(uniqueLinks, k)
	}

	return uniqueLinks
}

func notifyAboutUrlWithTime(url string, startTime time.Time, error bool, statusCode string) {
	// Construct notification
	executionTime := time.Now().Sub(startTime).Nanoseconds() / 1E+6
	message := ""

	if error {
		message += "\tERROR\t"
	} else {
		message += "\t" + statusCode + "\t"
	}
	message += "Parsed by " + strconv.FormatInt(executionTime, 10) + " ms\t"
	message += "url: " + url

	log.Print(message)
}
