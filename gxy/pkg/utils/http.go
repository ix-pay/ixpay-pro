package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// 全局复用的HTTP客户端，避免每次请求创建新客户端
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

func MakeHTTPRequest(method, url string, headers map[string]string, body interface{}) (*http.Response, error) {
	var requestBody []byte
	var err error

	if body != nil {
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	// 设置默认请求头
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义请求头
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 使用全局复用的HTTP客户端
	return httpClient.Do(req)
}

func ParseHTTPResponse(resp *http.Response, v interface{}) error {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if v != nil {
		return json.Unmarshal(data, v)
	}

	return nil
}
