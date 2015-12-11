package main

import (
       "errors"
	"flag"
	_ "fmt"
	api "github.com/whosonfirst/go-brooklynintegers-api"
	pool "github.com/whosonfirst/go-whosonfirst-pool"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Proxy struct {
	Client  *api.APIClient
	Pool    *pool.LIFOPool
	MinPool int64
}

func NewProxy(min_pool int64) *Proxy {

	client := api.NewAPIClient()
	pool := pool.NewLIFOPool()

	proxy := Proxy{
		Client:  client,
		Pool:    pool,
		MinPool: min_pool,
	}

	return &proxy
}

func (p *Proxy) Init() {

	wg := new(sync.WaitGroup)

	for i := 0; int64(i) < p.MinPool; i++ {

		wg.Add(1)

		go func(pr *Proxy) {
			defer wg.Done()
			pr.AddToPool()
		}(p)
	}

	wg.Wait()

	go func() {
		p.Monitor()
	}()
}

func (p *Proxy) Monitor() {

	for {

		if p.Pool.Length() < p.MinPool {

			wg := new(sync.WaitGroup)

			todo := p.MinPool - p.Pool.Length()

			for j := 0; int64(j) < todo; j++ {

				wg.Add(1)

				go func(pr *Proxy) {

					defer wg.Done()
					pr.AddToPool()
				}(p)
			}

			wg.Wait()
		}

		time.Sleep(1 * time.Second)
	}
}

func (p *Proxy) AddToPool() bool {

	i, err := p.GetInteger()

	if err != nil {
		return false
	}

	pi := pool.PoolInt{Int:i}

	p.Pool.Push(pi)
	return true
}

func (p *Proxy) GetInteger() (int64, error) {

	i, err := p.Client.CreateInteger()

	if err != nil {
		return 0, err
	}

	return i, nil
}

func (p *Proxy) Integer() (int64, error) {

	if p.Pool.Length() == 0 {
		return p.GetInteger()
	}

	i, ok := p.Pool.Pop()

	if !ok {
	   return 0, errors.New("Failed to pop")
	}

	return i.IntValue(), nil
}

func main() {

	var port = flag.Int("port", 8080, "Port to listen")
	var min = flag.Int("min", 10, "The minimum number of Brooklyn Integers to keep on hand at all times")
	var cors = flag.Bool("cors", false, "Enable CORS headers")

	flag.Parse()

	proxy := NewProxy(int64(*min))
	proxy.Init()

	handler := func(rsp http.ResponseWriter, r *http.Request) {

		i, err := proxy.Integer()

		if err != nil {
			http.Error(rsp, "Unknown placetype", http.StatusBadRequest)
		}

		if *cors {
			rsp.Header().Set("Access-Control-Allow-Origin", "*")
			return
		}

		io.WriteString(rsp, strconv.Itoa(int(i)))
	}

	http.HandleFunc("/", handler)

	str_port := ":" + strconv.Itoa(*port)
	err := http.ListenAndServe(str_port, nil)

	if err != nil {
		log.Fatal("Failed to start server, because %v\n", err)
	}

}
