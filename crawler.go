// portions of this code are copied from:
// https://jdanger.com/build-a-web-crawler-in-go.html
// https://schier.co/blog/2015/04/26/a-simple-web-scraper-in-go.html

// TODO: build keywordHighlight feature
// TODO: This will have to take starting url from web form
package crawler

import (
	"fmt"                   // output
	"golang.org/x/net/html" // parse html
	"log"                   // error logging
	"math/rand"             // for getting random numbers
	"net/http"              // really useful http package in stdlib
	"net/url"
	"time" // for seeding the random number
)

type Vertex struct {
	url              string
	keywordHighlight bool
	adjacentTo       []int
}

type Graph struct {
	associatedKeywords []string
	numVertices        int
	Vertices           []*Vertex
}

type Page struct {
	links   []string
	visited bool
}

const DEPTH = 30

func Crawl(startingUrl string) (*Graph, error) {
	// seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// this implements a depth first search for a hard coded depth
	pages := make(map[string]Page)
	stack := []string{startingUrl}

	visitCount := 0
	for len(stack) > 0 {
		if visitCount >= DEPTH {
			break
		}

		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1] // remove top of stack

		if pages[top].visited {
			continue
		}

		links, err := retrieveBody(top)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			continue
		}

		rand.Shuffle(len(links), func(i, j int) {
			links[i], links[j] = links[j], links[i]
		})
		pages[top] = Page{links: links, visited: true}
		visitCount++
		for _, link := range links {
			if pages[link].visited {
				continue
			}
			pages[link] = Page{visited: false}
			stack = append(stack, link)
		}
	}
	urlIndex := make(map[string]int)
	for pageUrl, _ := range pages {
		urlIndex[pageUrl] = len(urlIndex)
	}

	graph := new(Graph)
	for pageUrl, page := range pages {
		// create vertex and add to graph
		v := new(Vertex)
		v.url = pageUrl
		v.keywordHighlight = false

		for _, link := range page.links {
			v.adjacentTo = append(v.adjacentTo, urlIndex[link])
		}

		// update the graph
		graph.Vertices = append(graph.Vertices, v)
		graph.numVertices += 1

	}
	return graph, nil
}

// retrieveBody gets the html body at a url and return a slice of links in that body
func retrieveBody(pageUrl string) ([]string, error) {
	// in go, functions return two things, the return value and any errors
	// this double assignment takes the return value and error from .Get()
	// and assigns them to variables resp and err respectively
	resp, err := http.Get(pageUrl)

	if err != nil {
		return nil, fmt.Errorf("http transport error is: %v", err)
	}

	urlb, err := url.Parse(pageUrl)
	if err != nil {
		return nil, err
	}
	// .Get() opened a TCP connection to the url .Close() will close it
	// defer makes it so this function is not called until the function ends
	defer resp.Body.Close()

	// extract html body
	body := resp.Body

	// set up map for urls
	var foundUrl []string

	// parse html body for urls
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					urlp, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					foundUrl = append(foundUrl, urlb.ResolveReference(urlp).String())
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return foundUrl, err

	/*
		// iterate over tokens to find <a> tags
		z := html.NewTokenizer(body)
		for {
			tt := z.Next()
			switch {
			case tt == html.ErrorToken:
				// end of html body
			case tt == html.StartTagToken:
				t := z.Token()
				isAnchor := t.Data == "a"
				if isAnchor {
					// found a link
					for _, a := range t.Attr {
						if !(a.Key == "href") {
							continue
						}
						isFullLink := strings.Index(a.Val, "http") == 0
						if !isFullLink {
								continue
						}
						// if link has http, add to array
						foundUrl = append(foundUrl, a.Val)
					}
				}
			}
		}
	*/

}
