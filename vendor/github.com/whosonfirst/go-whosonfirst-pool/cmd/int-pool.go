package main

import (
       "fmt"
       "github.com/whosonfirst/go-whosonfirst-pool"
       "log"
)

func main() {

     p, err := pool.NewMemLIFOPool()

     if err != nil {
     	log.Fatal(err)
     }

     f := pool.NewIntItem(int64(123))

     p.Push(f)
     v, _ := p.Pop()

     fmt.Printf("%d", v.Int())
}
