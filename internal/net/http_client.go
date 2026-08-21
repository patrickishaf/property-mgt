package net

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient struct {
	BaseURL string
	Client  http.Client
	Headers map[string]string
}

func NewHTTPClient(baseURL string, headers map[string]string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client: http.Client{
			Timeout: 30 * time.Second,
		},
		Headers: headers,
	}
}

func (hc *HTTPClient) appendHeaders(req *http.Request) {
	for k, v := range hc.Headers {
		req.Header.Add(k, v)
	}
}

func (hc *HTTPClient) parseResponse(res *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if strings.Split(res.Status, "")[0] != "2" {
		body, _ := io.ReadAll(res.Body)
		err = errors.New(string(body))
		return nil, err
	}

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return responseBody, nil
}

func (hc *HTTPClient) Get(url string, params *url.Values) ([]byte, error) {
	if params != nil {
		url = url + "?" + params.Encode()
	}
	req, err := http.NewRequest(http.MethodGet, hc.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}
	hc.appendHeaders(req)
	res, err := hc.Client.Do(req)
	return hc.parseResponse(res, err)
}

func (hc *HTTPClient) Patch(url string, body any) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resBody := bytes.NewBuffer(jsonBody)
	req, err := http.NewRequest(http.MethodPatch, hc.BaseURL+url, resBody)
	if err != nil {
		return nil, err
	}
	hc.appendHeaders(req)
	res, err := hc.Client.Do(req)
	return hc.parseResponse(res, err)
}

func (hc *HTTPClient) Post(url string, body any) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resBody := bytes.NewBuffer(jsonBody)
	req, err := http.NewRequest(http.MethodPost, hc.BaseURL+url, resBody)
	if err != nil {
		return nil, err
	}
	hc.appendHeaders(req)
	res, err := hc.Client.Do(req)
	return hc.parseResponse(res, err)
}

func (hc *HTTPClient) SetHeader(key, value string) {
	hc.Headers[key] = value
}
