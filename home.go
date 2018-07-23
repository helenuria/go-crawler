// home.go runs the homepage,
// provides the user form,
// stores the response in
// in some format, crawls,
// then graphs the crawl.
// TODO: build keywordHighlight feature
// TODO add past starting urls using cookies/sessions
package main

import (
    "encoding/json"
	"fmt"                   // output
	"golang.org/x/net/html" // parse html
	"log"                   // error logging
	"math/rand"             // for getting random numbers
	"net/http"              // really useful http package in stdlib
	"net/url"
	"time"                  // for seeding the random number

	"flag"
	//"fmt"
	"html/template"
	//"log"
	//"net/http"

    "google.golang.org/appengine"
)

var (
	addr = flag.String("addr", ":80", "address for the server to listen on")
)

type CrawlSettings struct {
	Url     string // Start
	Keyword string // Optional
	Type    string // "B" or "D"
	BL      string // Breadth limit
	DL      string // Depth limit
}

type Graph struct {
	Nodes		string
	Links		string
	Success	bool
	CrawlUrl	string
}

type Vertex struct {
	Url              string
	KeywordHighlight bool
}

type Edge struct {
	Target string
	Source string
}

type Page struct {
	links   []string
	visited bool
}

const DEPTH = 30

func Crawl(startingUrl string) ([]byte, []byte, error) {
	// seed the random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// this implements a depth first search for a hard coded depth
	pages := make(map[string]Page)
	stack := []string{startingUrl}
	var Vertices []Vertex
	var Edges []Edge

	visitCount := 0
	for len(stack) > 0 {
		if visitCount >= DEPTH {
			break
		}

		// pop the stack
		top := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// if the link has already been visited, do not add to graph
		// this prevents loops
		if pages[top].visited {
			continue
		}

		// Get the links from the top link...
		links, err := retrieveBody(top)
		if err != nil {
			log.Printf("couldnt retrieve body: %v", err)
			continue
		}

		// ...and randomize the order (because we'll have to pop them in order)
		rand.Shuffle(len(links), func(i, j int) {
			links[i], links[j] = links[j], links[i]
		})

		// ...then mark the current link as visited.
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
			stack = append(stack, link)
		}
	}

	for pageUrl, page := range pages {
		// create vertex and add to graph
		v := new(Vertex)
		v.Url = pageUrl
		v.KeywordHighlight = false

		for _, link := range page.links {
			e := new(Edge)
			e.Target = link
			e.Source = pageUrl
			Edges = append(Edges, *e)
		}

		Vertices = append(Vertices, *v)
	}
	//fmt.Println("Vertices: ", Vertices, "\nEdges: ", Edges)
	vJson, err := json.Marshal(Vertices)
	eJson, err2 := json.Marshal(Edges)
	if err != nil && err2 != nil {
		log.Printf("couldnt parsse json: %v, %v", err, err2)
		return nil, nil, err
	}
	//fmt.Println("Vertices: ", string(vJson), "\nEdges: ", string(eJson))
	return vJson, eJson, nil
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
	fmt.Printf("%+v\n", crawl) // debug

	// Populate crawl graph.
	crawl_nodes, crawl_links, _ := Crawl(crawl.Url)
	// fmt.Println("vertices:\n", (crawl_nodes), "\nedges:\n", (crawl_links))
	json := Graph{Nodes: string(crawl_nodes), Links: string(crawl_links), Success: true, CrawlUrl: crawl.Url}
	// Render graph.
	tmpl.Execute(w, json)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}
        appengine.Main() // Starts the server to receive requests.
}
