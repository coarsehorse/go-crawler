package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	RESULTS_DIR   = "RESULTS"
	GORUTINES_NUM = 32
)

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
func writeDataToFile(file *os.File, data []byte) error {
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

type MyTest struct {
	alloe string `json:"aloe"`
	fork  int    `json:"fork"`
}

func filterSlice(slice []string, predicate func(string) bool) (filtered []string) {
	for _, s := range slice {
		if predicate(s) {
			filtered = append(filtered, s)
		}
	}
	return
}

func addFollowingSlash(str string) string {
	if chars := strings.Split(str, ""); chars[len(chars)-1] != "/" {
		return str + "/"
	}
	return str
}

func main() {
	//a := []int{1, 2, 3}
	//fmt.Println(a)
	//a = append(a, []int{3, 4, 5}...)
	//fmt.Println(a)
	//
	//v := 1000000000 / 1E9
	//
	//fmt.Println(v)

	//t := time.Now()
	//fmt.Println(t.Format("2-1-2006-15:04:05"))

	// Create unique backup file

	//f, err := createUniqResultFile("http://domain.du/asdasda", ".json")
	//
	//err = writeDataToFile(f, []byte("asdasd"))
	//_, err = f.WriteString("1111")
	//if err != nil {
	//	panic(err.Error())
	//}

	//tests := []MyTest{{"asdasd", 1}, {"Two", 3}}
	//
	//marshaled, err := json.MarshalIndent(tests, "", "\t")
	//if err != nil {
	//	panic(err.Error())
	//}
	//fmt.Println(marshaled)

	//url := "http://domain.com/"
	//
	//nextLevelLinks := []string {
	//	"mailto:asdasds.com", "tel:+123124", "script:javascript:void(0).sdfsdf", "javascript:;",
	//	"http://domain.com/page/1/", "http://example.com/page/1/",
	//}
	//// Filter out bad links(another domains, tel:, etc.)
	//domain := extractDomain(url)
	//domainPattern := `^(http|https):\/\/` + strings.Replace(domain, `.`, `\.`, -1) + `.*$`
	//r, err := regexp.Compile(domainPattern)
	//if err != nil {
	//	panic("Bad regexp constructed for links validation: \"" + domainPattern + "\"")
	//}
	//nextLevelLinks = filterSlice(nextLevelLinks, func(link string) bool {
	//	if link == "" || link == "#" {
	//		return false
	//	} else if strings.HasPrefix(link, "tel:") ||
	//		strings.HasPrefix(link, "mailto:") ||
	//		strings.Contains(link, "javascript:void(0)") ||
	//		strings.Contains(link, "javascript:;") {
	//
	//		return false
	//	}
	//	return r.MatchString(link) // check domain domainPattern
	//})
	//
	//fmt.Println(nextLevelLinks)

	url1 := "http://domain.com1/"
	url2 := "http://domain.com2"

	fmt.Println(addFollowingSlash(url1))
	fmt.Println(addFollowingSlash(url2))
}
