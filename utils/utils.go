package utils

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	RESULTS_DIR = "RESULTS"
)

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
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
func CreateUniqResultingFile(url string, ext string) (createdFile *os.File, err error) {
	domain := ExtractDomain(url)
	t := time.Now()
	date := t.Format("2-1-2006-15-04-05") // get datetime in string(dd-MM-YYYY-HH-mm-ss)
	curDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	_ = os.MkdirAll(filepath.Join(curDir, RESULTS_DIR), os.ModePerm) // create results dir

	fileName := filepath.Join(curDir, RESULTS_DIR, domain+"-"+date+ext) // absolute path for resulting file
	createdFile, err = os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return createdFile, nil
}

// Writes data to the already created file with error handling
// and closing opened file after writing
// Returns the possible error or nil on success
func WriteToFileAndClose(file *os.File, data []byte) error {
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

func UniqueStringSlice(initialSlice []string) []string {
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

func FilterSlice(slice []string, predicate func(string) bool) (filtered []string) {
	for _, s := range slice {
		if predicate(s) {
			filtered = append(filtered, s)
		}
	}
	return
}

/**
Adds the following slash if it not already exists in specified string
*/
func AddFollowingSlash(str string) string {
	if chars := strings.Split(str, ""); chars[len(chars)-1] != "/" {
		return str + "/"
	}
	return str
}

func ExtractDomain(url string) string {
	return strings.Split(url, "/")[2]
}
