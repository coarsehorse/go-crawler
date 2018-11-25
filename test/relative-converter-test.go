package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func addFollowingSlash(str string) string {
	if chars := strings.Split(str, ""); chars[len(chars)-1] != "/" {
		return str + "/"
	}
	return str
}

func extractDomain(url string) string {
	return strings.Split(url, "/")[2]
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
	// Test conversion

	// Initial data
	absLocat := "https://domain.com/one/two/three"
	fmt.Println("Initial data\n")
	fmt.Println("AbsoluteLocation: " + absLocat + "\n")

	// path/path1 case
	rel := "path/path1"
	res, err := extendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// ../path case
	rel = "../path"
	res, err = extendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("../path case")
	fmt.Println(res)
	fmt.Println()

	// /path/path1 case
	rel = "/path/path1"
	res, err = extendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("/path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// //path/path1 case
	rel = "//path/path1"
	res, err = extendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("//path/path1 case")
	fmt.Println(res)
	fmt.Println()

	// Already absolute like http://domain.com/path case
	rel = "http://domain.com/path"
	res, err = extendRelativeLink(rel, absLocat)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("http://domain.com/path case")
	fmt.Println(res)
	fmt.Println()
}
