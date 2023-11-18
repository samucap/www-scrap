package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gocolly/colly"
)

var baseURL string = "https://www.health.harvard.edu"
var userAgents []string

type pageInfo struct {
	Links []string
}

func isErr(err error) bool {
	if err != nil {
		fmt.Errorf("Big Trouble boss man, error: %v\n", err)
		return true
	}

	return false
}

func init() {
	f, err := os.Open("./data.txt")
	if isErr(err) {
		fmt.Errorf("shit")
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	chunks := []byte{}
	for scanner.Scan() {
		chunks = append(chunks, scanner.Bytes()...)
	}

	err = json.Unmarshal(chunks, &userAgents)
	isErr(err)
}

func main() {
	c := colly.NewCollector(
		colly.Async(true),
		colly.UserAgent(userAgents[rand.Intn(len(userAgents))]),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	p := &pageInfo{Links: []string{}}

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		fmt.Println("Visiting ", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Println("Visited: ", r.Request.URL)
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		//TODO filter next nodes:
		// - mailto
		// - by domain ?
		if link != "" {
			p.Links = append(p.Links, link)
			c.Visit(link)
		}
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("Finished: ", r.Request.URL)
	})

	c.WithTransport(&http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          50,
		IdleConnTimeout:       10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})

	c.Visit(baseURL)
	c.Wait()

	sort.Strings(p.Links)
	fmt.Printf("Crawl results %v\n", p)
}
