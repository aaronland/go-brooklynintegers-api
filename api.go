package api

import (
	"errors"
	"fmt"
	"github.com/aaronland/go-artisanal-integers"
	"github.com/tidwall/gjson"
	"go.uber.org/ratelimit"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"github.com/cenkalti/backoff/v4"
)

// this is basically just so we can preserve backwards compatibility
// even though the artisanalinteger.Client interface is the new new
// (20181210/thisisaaronland)

type BrooklynIntegersClient interface {
	CreateInteger() (int64, error)
	ExecuteMethod(string, *url.Values) (*APIResponse, error)
}

type APIClient struct {
	artisanalinteger.Client
	BrooklynIntegersClient // see above
	isa                    string
	http_client            *http.Client
	Scheme                 string
	Host                   string
	Endpoint               string
	rate_limiter           ratelimit.Limiter
}

type APIError struct {
	Code    int64
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

type APIResponse struct {
	raw []byte
}

func (rsp *APIResponse) Int() (int64, error) {

	ints := gjson.GetBytes(rsp.raw, "integers.0.integer")

	if !ints.Exists() {
		return -1, errors.New("Failed to generate any integers")
	}

	i := ints.Int()
	return i, nil
}

func (rsp *APIResponse) Stat() string {

	r := gjson.GetBytes(rsp.raw, "stat")

	if !r.Exists() {
		return ""
	}

	return r.String()
}

func (rsp *APIResponse) Ok() (bool, error) {

	stat := rsp.Stat()

	if stat == "ok" {
		return true, nil
	}

	return false, rsp.Error()
}

func (rsp *APIResponse) Error() error {

	c := gjson.GetBytes(rsp.raw, "error.code")
	m := gjson.GetBytes(rsp.raw, "error.message")

	if !c.Exists() {
		return errors.New("Failed to parse error code")
	}

	if !m.Exists() {
		return errors.New("Failed to parse error message")
	}

	err := APIError{
		Code:    c.Int(),
		Message: m.String(),
	}

	return &err
}

func NewAPIClient() artisanalinteger.Client {

	http_client := &http.Client{}
	rl := ratelimit.New(10)		// please make this configurable

	return &APIClient{
		Scheme:       "https",
		Host:         "api.brooklynintegers.com",
		Endpoint:     "rest/",
		http_client:  http_client,
		rate_limiter: rl,
	}
}

func (client *APIClient) CreateInteger() (int64, error) {
	return client.NextInt()
}

func (client *APIClient) NextInt() (int64, error) {

	params := url.Values{}
	method := "brooklyn.integers.create"

	var next_id int64

	cb := func() error {

		rsp, err := client.ExecuteMethod(method, &params)

		if err != nil {
			return err
		}

		i, err := rsp.Int()
		
		if err != nil {
			log.Println(err)
			return err
		}

		next_id = i
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	
	err := backoff.Retry(cb, bo)

	if err != nil {
		return -1, err
	}

	return next_id, nil
}

func (client *APIClient) ExecuteMethod(method string, params *url.Values) (*APIResponse, error) {

	client.rate_limiter.Take()

	url := client.Scheme + "://" + client.Host + "/" + client.Endpoint

	params.Set("method", method)

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = (*params).Encode()

	req.Header.Add("Accept-Encoding", "gzip")

	rsp, err := client.http_client.Do(req)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	r := APIResponse{
		raw: body,
	}

	return &r, nil
}
