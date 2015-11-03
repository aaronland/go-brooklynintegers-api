package main

import (
	"flag"
	_ "fmt"
	api "github.com/whosonfirst/go-brooklynintegers-api"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// https://github.com/SimonWaldherr/golang-examples/blob/2be89f3185aded00740a45a64e3c98855193b948/advanced/lifo.go

func NewPool() *Pool {
	return &Pool{mutex: &sync.Mutex{}}
}

type Pool struct {
	nodes []int64
	count int64
	mutex *sync.Mutex
}

func (pl *Pool) Length() int64 {
	return pl.count
}

func (pl *Pool) Push(n int64) {
	pl.nodes = append(pl.nodes[:pl.count], n)
	atomic.AddInt64(&pl.count, 1)
}

func (pl *Pool) Pop() int64 {

	if pl.count == 0 {
		return 0
	}

	pl.mutex.Lock()

	atomic.AddInt64(&pl.count, -1)
	i := pl.nodes[pl.count]

	pl.mutex.Unlock()
	return i
}

type Proxy struct {
	Client  *api.APIClient
	Pool    *Pool
	MinPool int64
}

func NewProxy(min_pool int64) *Proxy {

	client := api.NewAPIClient()
	pool := NewPool()

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

	p.Pool.Push(i)
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

	i := p.Pool.Pop()
	return i, nil
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
