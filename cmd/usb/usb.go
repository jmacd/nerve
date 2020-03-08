package main

import (
	"fmt"
	"log"

	"github.com/jmacd/nerve/lctlxl"
)

func main() {
	lc, err := lctlxl.Open()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Ready!", lc)
	select {}
}
