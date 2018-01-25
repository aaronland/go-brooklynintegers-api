package main

import (
	"bufio"
	"flag"
	"github.com/aaronland/go-brooklynintegers-api"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

func main() {

	var count = flag.Int("count", 1, "The number of Brooklyn Integers to mint.")

	flag.Parse()

	writers := []io.Writer{
		os.Stdout,
	}

	multi := io.MultiWriter(writers...)
	writer := bufio.NewWriter(multi)

	client := api.NewAPIClient()

	wg := new(sync.WaitGroup)
	mu := new(sync.Mutex)

	t := 10
	throttle := make(chan bool, t)

	for i := 0; i < t; i++ {

		throttle <- true
	}

	for i := 0; i < *count; i++ {

		wg.Add(1)

		go func(throttle chan bool) {

			<-throttle

			defer func() {
				throttle <- true
				wg.Done()
			}()

			i, err := client.CreateInteger()

			if err != nil {
				log.Fatal(err)
			}

			mu.Lock()

			str_i := strconv.Itoa(int(i))
			writer.WriteString(str_i + "\n")
			writer.Flush()

			mu.Unlock()

		}(throttle)

	}

	wg.Wait()
	os.Exit(0)
}
