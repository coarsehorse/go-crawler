package validator

import (
	"go-crawler/utils"
	"log"
	"regexp"
)

type Validator struct {
	Exceptions []*regexp.Regexp `json:"exceptions"`
	Allowances []*regexp.Regexp `json:"allowances"`
}

func (v Validator) IsValid(str string) (isValid bool) {
	// Check exception rules
	for _, exc := range v.Exceptions {
		if exc.MatchString(str) {
			// Check allowance rules
			for _, all := range v.Allowances {
				if all.MatchString(str) {
					return true
				}
			}
			return false
		}
	}

	// No exception rules
	return true
}

func NewValidator(exceptions []string, allowances []string) (newValidator Validator) {
	// Compile all the exception expressions
	var tempExceptions []*regexp.Regexp
	for _, e := range utils.RemoveEmptyStrings(utils.TrimArray(exceptions)) {
		r, err := regexp.Compile(e)
		if err == nil {
			tempExceptions = append(tempExceptions, r)
		} else {
			log.Println("[validator]\tBroken exception regexp: \"" + e + "\" skipping...")
		}
	}

	// Compile all the allowance expressions
	var tempAllowances []*regexp.Regexp
	for _, a := range utils.RemoveEmptyStrings(utils.TrimArray(allowances)) {
		r, err := regexp.Compile(a)
		if err == nil {
			tempAllowances = append(tempAllowances, r)
		} else {
			log.Println("[crawler]\tBroken allowance regexp: \"" + a + "\" skipping...")
		}
	}

	return Validator{tempExceptions, tempAllowances}
}
