package main

import (
	"errors"
	"flag"
	_ "fmt"
	api "github.com/whosonfirst/go-brooklynintegers-api"
	log "github.com/whosonfirst/go-whosonfirst-log"
	pool "github.com/whosonfirst/go-whosonfirst-pool"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Proxy struct {
	logger  *log.WOFLogger
	client  *api.APIClient
	pool    *pool.LIFOPool
	minpool int64
}

func NewProxy(min_pool int64, logger *log.WOFLogger) *Proxy {

	api_client := api.NewAPIClient()
	pool := pool.NewLIFOPool()

	proxy := Proxy{
		logger:  logger,
		client:  api_client,
		pool:    pool,
		minpool: min_pool,
	}

	return &proxy
}

func (p *Proxy) Init() {

	go p.RefillPool()

	go func() {
		p.Monitor()
	}()
}

func (p *Proxy) Monitor() {

	for {

		p.logger.Debug("pool: %d", p.pool.Length())

		if p.pool.Length() < p.minpool {
			p.RefillPool()
		}

		time.Sleep(1 * time.Second)
	}
}

func (p *Proxy) RefillPool() {

	wg := new(sync.WaitGroup)

	todo := p.minpool - p.pool.Length()

	for j := 0; int64(j) < todo; j++ {

		wg.Add(1)

		go func(pr *Proxy) {

			defer wg.Done()
			pr.AddToPool()
		}(p)
	}

	wg.Wait()
}

func (p *Proxy) AddToPool() bool {

	i, err := p.GetInteger()

	if err != nil {
		return false
	}

	pi := pool.PoolInt{Int: i}

	p.pool.Push(pi)
	return true
}

func (p *Proxy) GetInteger() (int64, error) {

	i, err := p.client.CreateInteger()

	if err != nil {
		return 0, err
	}

	return i, nil
}

func (p *Proxy) Integer() (int64, error) {

	if p.pool.Length() == 0 {
		return p.GetInteger()
	}

	i, ok := p.pool.Pop()

	if !ok {
		return 0, errors.New("Failed to pop")
	}

	return i.IntValue(), nil
}

func main() {

	var port = flag.Int("port", 8080, "Port to listen")
	var min = flag.Int("min", 1, "The minimum number of Brooklyn Integers to keep on hand at all times")
	var loglevel = flag.String("loglevel", "info", "Log level")
	var cors = flag.Bool("cors", false, "Enable CORS headers")

	flag.Parse()

	writer := io.MultiWriter(os.Stdout)

	logger := log.NewWOFLogger("[big-integer] ")
	logger.AddLogger(writer, *loglevel)

	proxy := NewProxy(int64(*min), logger)
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
		logger.Fatal("Failed to start server, because %v\n", err)
	}

}
