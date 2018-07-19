// home.go runs the homepage,
// provides the user form,
// stores the response in
// in some format, crawls,
// then graphs the crawl.
//TODO add cookies and past starting urls
package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/helenuria/go-crawler"
)

var (
	addr = flag.String("addr", ":80", "address for the server to listen on")
)

type Crawl struct {
	Url     string // Start
	Keyword string // Optional
	Type    string // "B" or "D"
	BL      string // Breadth limit
	DL      string // Depth limit
}

type Graph struct {
	//
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("forms.html"))

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	crawl := Crawl{
		Url:     r.FormValue("Url"),
		Keyword: r.FormValue("Keyword"),
		Type:    r.FormValue("Type"),
		BL:      r.FormValue("BL"),
		DL:      r.FormValue("DL"),
	}
	// crawl is now populated.
	fmt.Printf("%+v\n", crawl) // debug

	_, _ = crawler.Crawl(crawl.Url)

	// TODO
	/* Render graph
	input: graph (formatted)
	output: D3.js, graphs.html? */

	tmpl.Execute(w, struct{ Success bool }{true})
}

func exampleHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	if r.Method != http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}
	tmpl.Execute(w, struct { Success bool }{ true })
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
  http.HandleFunc("/example", exampleHandler)
	http.ListenAndServe(":80", nil)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

}
