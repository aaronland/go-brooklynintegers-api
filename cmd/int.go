package main

import (
	"fmt"
	api "github.com/whosonfirst/go-brooklynintegers-api"
)

func main() {

	client := api.NewAPIClient()
	i, _ := client.CreateInteger()

	fmt.Println(i)
}
