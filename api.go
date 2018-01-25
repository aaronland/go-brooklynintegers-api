package api

import (
       "encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
)

type APIClient struct {
	scheme   string
	isa      string
	Host     string
	Endpoint string
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

func (rsp *APIResponse) Stat() string {

     r := gjson.GetBytes(rsp.raw, "stat")

     if !r.Exists(){
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
	    Code: c.Int(),
	    Message: m.String(),
	}

	return &err
}

func (rsp *APIResponse) String() string {
     b, _ := json.Marshal(rsp.raw)
     return string(b)
}

func NewAPIClient() *APIClient {

	return &APIClient{
		scheme:   "http",
		Host:     "api.brooklynintegers.com",
		Endpoint: "rest/",
	}
}

func (client *APIClient) CreateInteger() (int64, error) {

	params := url.Values{}

	method := "brooklyn.integers.create"

	rsp, err := client.ExecuteMethod(method, &params)

	if err != nil {
		return 0, err
	}

	ints := gjson.GetBytes(rsp.raw, "integers.0")

	if !ints.Exists() {
		return -1, errors.New("Failed to generate any integers")
	}

	i := ints.Int()
	return i, nil
}

func (client *APIClient) ExecuteMethod(method string, params *url.Values) (*APIResponse, error) {

	url := client.scheme + "://" + client.Host + "/" + client.Endpoint

	params.Set("method", method)

	http_req, req_err := http.NewRequest("POST", url, nil)

	if req_err != nil {
		return nil, req_err
	}

	http_req.URL.RawQuery = (*params).Encode()

	http_req.Header.Add("Accept-Encoding", "gzip")

	http_client := &http.Client{}
	http_rsp, http_err := http_client.Do(http_req)

	if http_err != nil {
		return nil, http_err
	}

	defer http_rsp.Body.Close()

	http_body, io_err := ioutil.ReadAll(http_rsp.Body)

	if io_err != nil {
		return nil, io_err
	}

	r := APIResponse{
		raw: http_body,
	}

	return &r, nil
}
