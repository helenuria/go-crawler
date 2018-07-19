// home.go runs the homepage,
// provides the user form,
// stores the response in
// in some format, crawls,
// then graphs the crawl.
//TODO add past starting urls using cookies/sessions
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
	tmpl := template.Must(template.ParseFiles("index.html"))

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
	// Crawl settings is now populated.
	fmt.Printf("%+v\n", crawl) // debug

	// Populate crawl graph.
	_, _ = crawler.Crawl(crawl.Url)

	// Render graph.
	tmpl.Execute(w, struct{ Success bool }{true})
}

func main() {
	flag.Parse()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal(err)
	}

}
