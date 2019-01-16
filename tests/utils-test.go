package main

import (
	"go-crawler/utils"
	"log"
)

func main() {
	// RemoveEmptyStrings test
	inp := []string{"asd", "", "  ", "123", "a3d", "  d "}
	res := utils.RemoveEmptyStrings(inp)

	log.Println(res)
}
