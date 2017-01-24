package main

import (
	"flag"
	"fmt"
	"github.com/whosonfirst/go-brooklynintegers-api"
	"log"
)

func main() {

	var count = flag.Int("count", 1, "The number of Brooklyn Integers to mint")

	flag.Parse()

	client := api.NewAPIClient()

	for i := 0; i < *count; i++ {

		i, err := client.CreateInteger()

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(i)
	}
}
