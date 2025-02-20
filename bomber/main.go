package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Req struct {
	url   string
	body  string
	rType string
}

type Resp struct {
	url  string
	code int
	body string
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 10; i++ {
		go func() {
			for {
				_, err := request(&Req{
					url:   "http://127.0.0.1:8080/",
					body:  "",
					rType: "get",
				})
				if err != nil {
					log.Printf("request error: %v", err)
				}
				time.Sleep(time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

func request(req *Req) (*Resp, error) {
	var resp *http.Response
	var err error
	switch req.rType {
	case "get":
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		resp, err = client.Get(req.url)
	case "post":
		resp, err = http.Post(req.url, "text/plain; charset=utf-8", strings.NewReader(req.body))
	}
	if err != nil {
		return nil, fmt.Errorf("%s request error: %v", req.url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s request error: %v", req.url, err)
	}

	dataResp := Resp{
		url:  req.url,
		code: resp.StatusCode,
		body: string(body),
	}

	return &dataResp, nil
}
