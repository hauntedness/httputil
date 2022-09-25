package httputil

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36 Edg/93.0.961.38"
	ContentJson      = "application/json; charset=UTF-8"
)

type H map[string]string

var client = &http.Client{}

var retryQuata int = 1

func Get(url string, headers H) ([]byte, error) {
	return Request(http.MethodGet, url, nil, headers)
}

func Post(url string, body io.Reader, headers H) ([]byte, error) {
	return Request(http.MethodPost, url, body, headers)
}

// send request and retry 3 times
func Request(method string, url string, body io.Reader, headers H) ([]byte, error) {
	var err error
	// fmt.Println("going to: ", url)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		log.Println("ERROR: ", err)
	}
	// default setting
	req.Header.Set("Content-Type", ContentJson)
	req.Header.Set("User-Agent", DefaultUserAgent)

	// replace with input headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	var resp *http.Response
	for retries := 0; retries < retryQuata; {
		resp, err = client.Do(req)
		if err != nil {
			req.Body = io.NopCloser(body)
			req.GetBody = func() (io.ReadCloser, error) {
				return io.NopCloser(body), nil
			}
			retries = retries + 1
		} else {
			break
		}
	}
	if resp == nil {
		return nil, err
	}
	var res []byte
	res, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	resp.Body.Close()
	return res, err
}

func SetRetry(retryTimes int) {
	retryQuata = retryTimes + 1
}

var once = sync.Once{}

func SetProxy(proxyUrl string) {
	once.Do(func() {
		// Transport caches connections for future re-use,
		// if call SetProxy too many times, the connection pool will no longer make sense
		transport := &http.Transport{}
		transport.Proxy = func(_ *http.Request) (*url.URL, error) {
			if proxyUrl == "" {
				proxyUrl = "http://127.0.0.1:58309"
			}
			return url.Parse(proxyUrl)
		}
		client.Transport = transport
	})
}
