package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	limit   = 20
	postReq = "http://192.168.75.1:8080/"
	file    = "urls.txt"
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
	urls := make([]string, 0, 1000)
	chReq := make(chan *Req)
	chResp := make(chan *Resp)
	chErr := make(chan error)
	var wg sync.WaitGroup

	file, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, string(scanner.Bytes()))
	}
	file.Close()

	for i := 0; i < limit; i++ {
		go func() {
			requester(&wg, chReq, chResp, chErr)
		}()
	}
	go responder(chResp, chErr)

	for _, url := range urls {
		resp, err := request(&Req{url: postReq, rType: "post", body: url})
		if err != nil {
			log.Printf("url %s request error: %v", url, err)
			continue
		}
		wg.Add(10)
		go func() {
			for i := 0; i < 10; i++ {
				chReq <- &Req{url: resp.body, rType: "get"}
			}
		}()
	}

	wg.Wait()
}

func responder(chResp chan *Resp, chErr chan error) {
	for {
		select {
		case resp := <-chResp:
			if resp.code != 201 && resp.code != 307 {
				log.Printf("url %s, wrong code: %d", resp.url, resp.code)
			}
		case err := <-chErr:
			log.Println(err)
		}
	}
}

func requester(wg *sync.WaitGroup, chReq chan *Req, chResp chan *Resp, chErr chan error) {
	for req := range chReq {
		resp, err := request(req)
		if err != nil {
			chErr <- err
		} else {
			chResp <- resp
		}
		wg.Done()
	}
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
