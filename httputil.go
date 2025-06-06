// package httputil used for making temporary or one time request, it serves for simple task only.
package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36 Edg/93.0.961.38"
	ContentJson      = "application/json; charset=UTF-8"
)

type H map[string]string

var client = &http.Client{}

func Get(url string, headers H) (data []byte, err error) {
	return Request(http.MethodGet, url, nil, headers)
}

func Post(url string, body io.Reader, headers H) (data []byte, err error) {
	return Request(http.MethodPost, url, body, headers)
}

func GetJson[T any](url string, queryObject any, headers H) (value *T, err error) {
	return Json[T](http.MethodGet, url, queryObject, headers)
}

func PostJson[T any](url string, queryObject any, headers H) (value *T, err error) {
	return Json[T](http.MethodPost, url, queryObject, headers)
}

func Json[T any](method string, url string, queryObject any, headers H) (value *T, err error) {
	var data []byte
	if queryObject != nil {
		data, err = json.Marshal(queryObject)
		if err != nil {
			return nil, err
		}
	}
	reader := bytes.NewReader(data)
	if headers == nil {
		headers = make(H)
	}
	headers["Content-Type"] = ContentJson
	data, err = Request(method, url, reader, headers)
	if err != nil {
		return nil, err
	}
	t := new(T)
	err = json.Unmarshal(data, t)
	if err != nil {
		if len(data) >= 1024 {
			data = data[0:1025]
			copy(data[1022:1025], "...")
		}
		return nil, fmt.Errorf("json.Unmarshal, data: %s, err: %w", string(data), err)
	}
	return t, nil
}

// RequestAndWriteTo send request and copy response body to dst
// it return [ErrContenthCopyLength] if the copied number of bytes doesn't match ContentLength
func RequestAndWriteTo(dst io.Writer, method string, url string, body io.Reader, headers H) (err error) {
	// fmt.Println("going to: ", url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	// default setting
	req.Header.Set("User-Agent", DefaultUserAgent)

	// replace with input headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(dst, resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return &StatusError{StatusCode: resp.StatusCode}
	}
	return nil
}

// Download create file on filepath send request and copy response body to the file
// it return [ErrContenthCopyLength] if the copied number of bytes doesn't match ContentLength
func Download(filepath string, method string, url string, body io.Reader, headers H) (err error) {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	return RequestAndWriteTo(file, method, url, body, headers)
}

// send request
func Request(method string, url string, body io.Reader, headers H) (data []byte, err error) {
	// fmt.Println("going to: ", url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	// default setting
	req.Header.Set("User-Agent", DefaultUserAgent)

	// replace with input headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &StatusError{StatusCode: resp.StatusCode}
	}
	return res, err
}

func SetProxy(proxyUrl string) {
	// Transport caches connections for future re-use,
	// if call SetProxy too many times, the connection pool will no longer make sense
	transport := &http.Transport{}
	transport.Proxy = func(_ *http.Request) (*url.URL, error) {
		return url.Parse(proxyUrl)
	}
	client.Transport = transport
}
