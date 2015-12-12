prep:
	if test -d pkg; then rm -rf pkg; fi

self:   prep
	if test -d src/github.com/whosonfirst/go-brooklynintegers-api; then rm -rf src/github.com/whosonfirst/go-brooklynintegers-api; fi
	mkdir -p src/github.com/whosonfirst/go-brooklynintegers-api
	cp api.go src/github.com/whosonfirst/go-brooklynintegers-api/

fmt:
	go fmt *.go
	go fmt cmd/*.go

deps:
	go get "github.com/jeffail/gabs"
	go get "github.com/whosonfirst/go-whosonfirst-pool"
	go get "github.com/whosonfirst/go-whosonfirst-log"

bin:	int proxy

int:	self
	go build -o bin/int cmd/int.go

proxy:	self
	go build -o bin/proxy-server cmd/proxy-server.go

