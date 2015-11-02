package main

import (
	"fmt"
	api "github.com/whosonfirst/go-brooklynintegers-api"
	"log"
	"sync"
)

// https://github.com/SimonWaldherr/golang-examples/blob/2be89f3185aded00740a45a64e3c98855193b948/advanced/lifo.go

func NewPool() *Pool {
	return &Pool{}
}

type Pool struct {
	nodes []int
	count int
}

func (pl *Pool) Length() int {
	return pl.count
}

func (pl *Pool) Push(n int) {
	pl.nodes = append(pl.nodes[:pl.count], n)

	fmt.Printf("COUNT %d\n", len(pl.nodes))
	pl.count++
}

func (pl *Pool) Pop() int {
	if pl.count == 0 {
		return 0
	}
	pl.count--
	return pl.nodes[pl.count]
}

type Proxy struct {
	Client  *api.APIClient
	Pool    *Pool
	MinPool int
	MaxPool int
}

func NewProxy(min_pool int, max_pool int) *Proxy {

	client := api.NewAPIClient()
	pool := NewPool()

	proxy := Proxy{
		Client:  client,
		Pool:    pool,
		MinPool: min_pool,
		MaxPool: max_pool,
	}

	return &proxy
}

func (p Proxy) Init() {

	wg := new(sync.WaitGroup)

	for i := 0; i < p.MinPool; i++ {

		wg.Add(1)

		go func() {
			defer wg.Done()
			p.AddToPool()
		}()
	}

	wg.Wait()

}

func (p Proxy) Monitor() {

	for {

		for p.Pool.Length() < p.MinPool {

			go func() {
				p.AddToPool()
			}()
		}
	}
}

func (p Proxy) AddToPool() bool {

	i, err := p.GetInteger()

	if err != nil {
		log.Fatal(err)
		return false
	}

	fmt.Println(i)
	p.Pool.Push(i)

	return true
}

func (p Proxy) GetInteger() (int, error) {

	i, err := p.Client.CreateInteger()

	if err != nil {
		return 0, err
	}

	return i, nil
}

func (p Proxy) Integer() (int, error) {

	if p.Pool.Length() == 0 {
		return p.GetInteger()
	}

	// TO DO : LOCK AND UNLOCK ME

	i := p.Pool.Pop()
	return i, nil
}

func main() {

	proxy := NewProxy(10, 15)
	proxy.Init()

	go proxy.Monitor()

	fmt.Println(proxy.Pool.Length())

	i, _ := proxy.Integer()
	fmt.Println(i)

	fmt.Println(proxy.Pool.Length())

}
