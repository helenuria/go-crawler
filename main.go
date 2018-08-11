// main.go uses th Google Cloud
// App Engine to host the crawler app.
// It gets the crawl settings by form,
// crawls, and graphs the crawl with D3.js.
// TODO build keywordHighlight feature
package main

import (
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
}

type Edge struct {
	Target int
	Source int
}

type Page struct {
	links   []string
	visited bool
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

func BreadthFirst(startingUrl string, r *http.Request, limit int) map[string]Page {
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
		links, err := retrieveBody(top, r)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			if levelSize == 0 {
				levelSize = allLinks.length()
				depthCount++
			}
			continue
		}

		// mark the current link as visited.
		pages[top] = Page{links: links, visited: true}

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

		if levelSize == 0 {
			levelSize = allLinks.length()
			depthCount++
		}
	}

	return pages
}

func DepthFirst(startingUrl string, r *http.Request, limit int) map[string]Page {
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
		links, err := retrieveBody(top, r)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			continue
		}

		// shuffle the links before adding them to the stacj- to randomize depth crawl
		Shuffle(links)

		// mark the current link as visited.
		pages[top] = Page{links: links, visited: true}
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
	}

	return pages
}

func Crawl(startingUrl string, r *http.Request, crawlType string, BL string, DL string) (
	[]Vertex, []Edge, error) {

	pages := make(map[string]Page)
	if crawlType == "B" {
		breadthLimit, err := strconv.Atoi(BL)
		if err != nil {
			fmt.Errorf("could not parse limit: %s", BL)
			return nil, nil, err
		}
		pages = BreadthFirst(startingUrl, r, breadthLimit)
	} else if crawlType == "D" {
		depthLimit, err := strconv.Atoi(DL)
		if err != nil {
			fmt.Errorf("could not parse limit: %s", BL)
			return nil, nil, err
		}
		pages = DepthFirst(startingUrl, r, depthLimit)
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
		v.KeywordHighlight = false

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
func retrieveBody(pageUrl string, r *http.Request) ([]string, error) {
	// Set up App Engine client, https://cloud.google.com/appengine/docs/go/urlfetch/
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	resp, err := client.Get(pageUrl)
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

	// set up slice for urls
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
}

// Save url and keyword history with cookies. 
func bake(crawl *CrawlSettings, w http.ResponseWriter, r *http.Request) (err error) {
	var sep string = " : "
	if _, err := r.Cookie("urlHistory"); err != nil {
		c := http.Cookie{Name: "urlHistory", Value: crawl.Url, Path: "/"}
		http.SetCookie(w, &c)
	} else {
		c1, _ := r.Cookie("urlHistory")
		v := fmt.Sprintf("%s%s%s", c1.Value, sep, crawl.Url)
		_ = trimDuplicates(&v, sep)
		c2 := http.Cookie{Name: "urlHistory", Value: v, Path: "/"}
		http.SetCookie(w, &c2)
	}
	if _, err := r.Cookie("keywordHistory"); err != nil {
		c := http.Cookie{Name: "keywordHistory", Value: crawl.Keyword, Path: "/"}
		http.SetCookie(w, &c)
	} else {
		c1, _ := r.Cookie("keywordHistory")
		v := fmt.Sprintf("%s%s%s", c1.Value, sep, crawl.Keyword)
		_ = trimDuplicates(&v, sep)
		c2 := http.Cookie{Name: "keywordHistory", Value: v, Path: "/"}
		http.SetCookie(w, &c2)
	}
	return nil
}

// Remove duplicate values in cookie value string. 
// https://www.dotnetperls.com/duplicates-go
func trimDuplicates(s *string, sep string) (err error) {
	vals := strings.Split(*s, sep)
	encountered := make(map[string]bool)
	result := []string{}
	for i := range vals {
		if encountered[vals[i]] == true {
			// Don't add duplicate.
		} else {
			encountered[vals[i]] = true
			result = append(result, vals[i])
		}
	}
	*s = strings.Join(result, sep)
	return nil
}

// Responds to an HTTP request.
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
	//fmt.Printf("%+v\n", crawl)

	if err := bake(&crawl, w, r); err != nil {
		fmt.Println(err)
	}

	// Populate crawl graph.
	crawl_nodes, crawl_links, _ := Crawl(crawl.Url, r, crawl.Type, crawl.BL, crawl.DL)
	// fmt.Println("vertices:\n", (crawl_nodes), "\nedges:\n", (crawl_links))
	json := Graph{Nodes: crawl_nodes, Links: crawl_links, Success: true, CrawlUrl: crawl.Url}
	// Render graph.
	tmpl.Execute(w, json)
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	flag.Parse()
	http.HandleFunc("/", handler)
	appengine.Main() // Starts the server to receive requests.
}
