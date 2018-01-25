package api

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	_ "log"
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
		return -1, err
	}

	return rsp.Int()
}

func (client *APIClient) ExecuteMethod(method string, params *url.Values) (*APIResponse, error) {

	url := client.scheme + "://" + client.Host + "/" + client.Endpoint

	params.Set("method", method)

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = (*params).Encode()

	req.Header.Add("Accept-Encoding", "gzip")

	cl := &http.Client{}
	rsp, err := cl.Do(req)

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
