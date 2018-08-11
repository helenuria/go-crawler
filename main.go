// main.go uses th Google Cloud
// App Engine to host the crawler app.
// It gets the crawl settings by form,
// crawls, and graphs the crawl with D3.js.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"
)

type CrawlSettings struct {
	Url     string // Start
	Keyword string // Optional
	Type    string // "B" or "D"
	BL      string // Breadth limit
	DL      string // Depth limit
}

type Graph struct {
	Nodes    []Vertex
	Links    []Edge
	Success  bool
	CrawlUrl string
}

type Vertex struct {
	Url              string
	KeywordHighlight bool
	Title            string
}

type Edge struct {
	Target int `json:"target"`
	Source int `json:"source"`
}

type Page struct {
	links   []string
	visited bool
	hasKey  bool
	title   string
}

// Shuffle function borrowed from
// https://www.calhoun.io/how-to-shuffle-arrays-and-slices-in-go/
func Shuffle(vals []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for len(vals) > 0 {
		n := len(vals)
		randIndex := r.Intn(n)
		vals[n-1], vals[randIndex] = vals[randIndex], vals[n-1]
		vals = vals[:n-1]
	}
}

type traverser interface {
	push(link string)
	pop() string
	length() int
}

type stack []string

func (s *stack) push(link string) {
	*s = append(*s, link)
}

func (s *stack) pop() string {
	top := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return top
}

func (s *stack) length() int {
	return len(*s)
}

type queue []string

func (q *queue) push(link string) {
	*q = append(*q, link)
}

func (q *queue) pop() string {
	top := (*q)[0]
	*q = (*q)[1:]
	return top
}

func (q *queue) length() int {
	return len(*q)
}

func BreadthFirst(startingUrl string, r *http.Request, limit int, keyWord string) map[string]Page {
	// pages will hold all the info we need to pass to the graph
	pages := make(map[string]Page)

	// allLinks is a queue to hold the URLs (as strings) we find
	var allLinks traverser
	temp := queue(nil)
	allLinks = &temp
	allLinks.push(startingUrl)

	depthCount := 0
	levelSize := allLinks.length()
	for allLinks.length() > 0 {
		if depthCount >= limit {
			break
		}

		// pop from the queue
		top := allLinks.pop()
		levelSize--

		// if the link has already been visited, do not add to graph
		// this prevents loops
		if pages[top].visited {
			if levelSize == 0 {
				levelSize = allLinks.length()
				depthCount++
			}
			continue
		}

		// visit the first url and find all the urls it links to
		title, links, foundKey, err := retrieveBody(top, r, keyWord)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			if levelSize == 0 {
				levelSize = allLinks.length()
				depthCount++
			}
			continue
		}

		// mark the current link as visited.
		if foundKey {
			pages[top] = Page{links: links, visited: true, hasKey: true, title: title}
		} else {
			pages[top] = Page{links: links, visited: true, hasKey: false, title: title}
		}

		// add the new links to the queue
		// this way, the next link we pop will be a sibling
		// until we run out of siblings, then it will be the
		// first child of the first sibling
		for _, link := range links {
			if pages[link].visited {
				continue
			}
			pages[link] = Page{visited: false}
			allLinks.push(link)
		}

		if foundKey {
			break
		}

		if levelSize == 0 {
			levelSize = allLinks.length()
			depthCount++
		}
	}

	return pages
}

func DepthFirst(startingUrl string, r *http.Request, limit int, keyWord string) map[string]Page {
	// pages will hold all the urls we'll format for the graph
	pages := make(map[string]Page)

	// allLinks is a stack containing all the links (URL strings) we find
	var allLinks traverser
	temp := stack(nil)
	allLinks = &temp
	allLinks.push(startingUrl)

	visitCount := 0
	for allLinks.length() > 0 {
		if visitCount >= limit {
			break
		}

		// pop the stack
		top := allLinks.pop()

		// if the link has already been visited, do not add to graph
		// this prevents loops
		if pages[top].visited {
			continue
		}

		// visit the url at the top and get all the urls it links to
		pageTitle, links, foundKey, err := retrieveBody(top, r, keyWord)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			continue
		}

		// shuffle the links before adding them to the stacj- to randomize depth crawl
		Shuffle(links)

		// mark the current link as visited.
		if foundKey {
			pages[top] = Page{links: links, visited: true, hasKey: true, title: pageTitle}
		} else {
			pages[top] = Page{links: links, visited: true, hasKey: false, title: pageTitle}
		}
		visitCount++

		// push the new links to the stack
		// this way, the next link we pop will be a child of the current link
		// unless there are no children, in which case we'll get a sibling link
		for _, link := range links {
			if pages[link].visited {
				continue
			}
			pages[link] = Page{visited: false}
			allLinks.push(link)
		}
		if foundKey {
			break
		}
	}

	return pages
}

func Crawl(startingUrl string, r *http.Request, crawlType string, BL string, DL string, keyWord string) (
	[]Vertex, []Edge, error) {

	pages := make(map[string]Page)
	if crawlType == "B" {
		breadthLimit, err := strconv.Atoi(BL)
		if err != nil {
			fmt.Errorf("could not parse limit: %s", BL)
			return nil, nil, err
		}
		pages = BreadthFirst(startingUrl, r, breadthLimit, keyWord)
	} else if crawlType == "D" {
		depthLimit, err := strconv.Atoi(DL)
		if err != nil {
			fmt.Errorf("could not parse limit: %s", BL)
			return nil, nil, err
		}
		pages = DepthFirst(startingUrl, r, depthLimit, keyWord)
	} else {
		return nil, nil, fmt.Errorf("incorrect crawl type parameter: %s", crawlType)
	}

	var Vertices []Vertex
	var Edges []Edge

	i := 0
	idMap := make(map[string]int)
	for pageUrl := range pages {
		// create vertex and add to graph
		v := new(Vertex)
		v.Url = pageUrl
		if pages[pageUrl].hasKey {
			v.KeywordHighlight = true
		} else {
			v.KeywordHighlight = false
		}
		pageTitle := pages[pageUrl].title
		v.Title = pageTitle

		Vertices = append(Vertices, *v)
		idMap[pageUrl] = i
		i++
	}

	for pageUrl, page := range pages {
		for _, link := range page.links {
			e := new(Edge)
			e.Target = idMap[link]
			e.Source = idMap[pageUrl]
			Edges = append(Edges, *e)
		}
	}

	return Vertices, Edges, nil
}

// retrieveBody gets the html body at a url and return a slice of links in that body
func retrieveBody(pageUrl string, r *http.Request, keyWord string) (string, []string, bool, error) {
	// Set up App Engine client, https://cloud.google.com/appengine/docs/go/urlfetch/
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	resp, err := client.Get(pageUrl)
	if err != nil {
		return "", nil, false, fmt.Errorf("http transport error is: %v", err)
	}

	urlb, err := url.Parse(pageUrl)
	if err != nil {
		return "", nil, false, err
	}
	// .Get() opened a TCP connection to the url .Close() will close it
	// defer makes it so this function is not called until the function ends
	defer resp.Body.Close()

	// extract html body
	body := resp.Body

	// set up slice for urls
	var foundUrl []string

	// parse html body for urls
	doc, err := html.Parse(body)
	if err != nil {
		return "", nil, false, err
	}

	var pattern *regexp.Regexp
	if keyWord != "" {
		pattern, err = regexp.Compile("(?i)\\b" + keyWord + "\\b")
		if err != nil {
			return "", nil, false, err
		}
	}

	var titleText string

	var f func(*html.Node) bool
	f = func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "title" {
			titleText = getText(n)
		}
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
		if keyWord != "" {
			if n.Type == html.TextNode {
				if strings.TrimSpace(n.Data) != "" {
					if pattern.MatchString(n.Data) {
						return true
					}
				}
			}
			if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "noscript") {
				return false
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if f(c) {
				return true
			}
		}
		return false
	}
	if f(doc) {
		return titleText, foundUrl, true, nil
	}

	return titleText, foundUrl, false, nil
}

func getText(body *html.Node) string {
	var foundString bytes.Buffer
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			if strings.TrimSpace(n.Data) != "" {
				foundString.WriteString(n.Data)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)
	return foundString.String()
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	crawl := CrawlSettings{
		Url:     r.FormValue("Url"),
		Keyword: r.FormValue("Keyword"),
		Type:    r.FormValue("Type"),
		BL:      r.FormValue("BL"),
		DL:      r.FormValue("DL"),
	}
	// Crawl settings is now populated.
	//fmt.Printf("%+v\n", crawl) // debug

	// Cookies.
	cookie := http.Cookie{Name: "urlHistory", Value: crawl.Url, Path: "/"}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{Name: "keywordHistory", Value: crawl.Keyword, Path: "/"}
	http.SetCookie(w, &cookie)

	// Populate crawl graph.
	crawlNodes, crawlLinks, _ := Crawl(crawl.Url, r, crawl.Type, crawl.BL, crawl.DL, crawl.Keyword)
	if crawlLinks == nil {
		crawlLinks = []Edge{}
	}
	// fmt.Println("vertices:\n", (crawlNodes), "\nedges:\n", (crawLinks))
	json := Graph{Nodes: crawlNodes, Links: crawlLinks, Success: true, CrawlUrl: crawl.Url}
	// Render graph.
	tmpl.Execute(w, json)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	http.HandleFunc("/", handler)
	appengine.Main() // Starts the server to receive requests.
}
