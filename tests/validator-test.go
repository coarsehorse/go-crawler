package main

import (
	"go-crawler/validator"
	"log"
)

func test(v validator.Validator, inp []string) {
	for _, s := range inp {
		log.Println("Validator:", v, "input:", s, "isValid: ", v.IsValid(s))
	}
}

func main() {
	var exceptions []string
	var allowances []string
	var v validator.Validator
	var testInput []string

	// Case 1 - Matched with one exception - one false
	exceptions = []string{"^ftp://.*$"}
	allowances = []string{"", " "}
	v = validator.NewValidator(exceptions, allowances)
	testInput = []string{"http://a.com", "ftp://f.com", "http://b.com"}
	test(v, testInput)
	log.Println()

	// Case 2 - Matched with one exception and one allowance - all true
	exceptions = []string{"^ftp://.*$"}
	allowances = []string{"ftp"}
	v = validator.NewValidator(exceptions, allowances)
	testInput = []string{"http://a.com", "ftp://f.com", "http://b.com"}
	test(v, testInput)
	log.Println()

	// Case 3 - No expressions - all true
	exceptions = []string{}
	allowances = []string{}
	v = validator.NewValidator(exceptions, allowances)
	testInput = []string{"http://a.com", "ftp://f.com", "http://b.com"}
	test(v, testInput)
	log.Println()

	// Case 4 - Real links with ? - two false
	exceptions = []string{`^.*[?].*$`}
	allowances = []string{`instagram.com[?]`}
	v = validator.NewValidator(exceptions, allowances)
	testInput = []string{"http://a.com?id=5", "http://b.com?id=1", "http://instagram.com?id=5"}
	test(v, testInput)
	log.Println()
}
