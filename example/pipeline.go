package main

import (
	"fmt"
	"net/http"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Set your email here to include in the User-Agent string
var email = "youremail@gmail.com"
var urls = []string{
	"https://news.ycombinator.com/",
	"https://httpbin.org/get",
	"https://www.reddit.com/",
	"https://en.wikipedia.org",
}

func respGen(urls []string) <-chan *http.Response {
	out := make(chan *http.Response)
	go func() {
		for _, url := range urls {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				panic(err)
			}
			req.Header.Set("user-agent", "testBot("+email+")")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			out <- resp
		}
		close(out)
	}()
	return out
}

func rootGen(in <-chan *http.Response) <-chan *html.Node {
	out := make(chan *html.Node)
	go func() {
		for resp := range in {
			root, err := html.Parse(resp.Body)
			if err != nil {
				panic(err)
			}
			out <- root
		}
		close(out)
	}()
	return out
}

func titleGen(in <-chan *html.Node) <-chan string {
	out := make(chan string)
	go func() {
		for root := range in {
			title, ok := scrape.Find(root, scrape.ByTag(atom.Title))
			if ok {
				out <- scrape.Text(title)
			}
		}
		close(out)
	}()
	return out
}

func main() {
	// Set up the pipeline to consume the back-to-back output
	// ending with the final stage of printing the title of
	// each web page in the main go routine.
	for title := range titleGen(rootGen(respGen(urls))) {
		fmt.Println(title)
	}
}
