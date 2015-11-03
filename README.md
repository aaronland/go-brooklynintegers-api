# go-brooklynintegers-api

Go language API library for the Brooklyn Integers API.

## Usage

## Simple

```
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
```

## Less simple

```
import (
       "fmt"
       api "github.com/whosonfirst/go-brooklynintegers-api"
)

client := api.NewAPIClient()

method := "brooklyn.integers.create"
params := url.Values{}

rsp, err := client.ExecuteMethod(method, &params)

if err != nil {
	return 0, err
}

ints, _ := rsp.Parsed.S("integers").Children()

if len(ints) == 0 {
	return 0, errors.New("Failed to generate any integers")
}

first := ints[0]

f, ok := first.Path("integer").Data().(float64)

if !ok {
	return 0, errors.New("Failed to parse response")
}

i := int64(f)
return i, nil
```

## HTTP Ponies

### proxy-server

```
$> ./bin/proxy-server -h
Usage of ./bin/proxy-server:
  -cors
	Enable CORS headers
  -min int
        (default 10)
  -port int
    	Port to listen (default 8080)
```

As in:

```
$> ./bin/proxy-server
```

And then:

```
$> curl http://localhost:8080
404573621
```

The `proxy-server` application needs more better logging and general reporting, still.

## See also

* http://brooklynintegers.com/
* http://brooklynintegers.com/api
