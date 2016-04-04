prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-brooklynintegers-api; then rm -rf src/github.com/whosonfirst/go-brooklynintegers-api; fi
	mkdir -p src/github.com/whosonfirst/go-brooklynintegers-api
	cp api.go src/github.com/whosonfirst/go-brooklynintegers-api/

rmdeps:
	if test -d src; then rm -rf src; fi 

build:	rmdeps deps fmt bin

fmt:
	go fmt *.go
	go fmt cmd/*.go

deps:
	@GOPATH=$(shell pwd) go get "github.com/jeffail/gabs"
	@GOPATH=$(shell pwd) go get "github.com/whosonfirst/go-whosonfirst-pool"
	@GOPATH=$(shell pwd) go get "github.com/whosonfirst/go-whosonfirst-log"

bin:	int proxy

int:	self
	@GOPATH=$(shell pwd) go build -o bin/int cmd/int.go

proxy:	self
	@GOPATH=$(shell pwd) go build -o bin/proxy-server cmd/proxy-server.go

