package main

import (
	"go-crawler/utils"
	"log"
)

func test(testName string, before interface{}, after interface{}) {
	log.Println(testName, "test:")
	log.Println("Before: ", before)
	log.Println("After: ", after)
	log.Println()
}

func main() {
	// RemoveEmptyStrings test
	inp1 := []string{"asd", "", "  ", "123", "a3d", "  d "}
	test("RemoveEmptyStrings", inp1, utils.RemoveEmptyStrings(inp1))

	// TrimArray test
	inp2 := []string{" asd 		", "zxc 1 asd", "\r\n asd p"}
	test("TrimArray", inp2, utils.TrimArray(inp2))
}
