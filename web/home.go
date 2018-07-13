// home.go runs the homepage,
// provides the user form,
// stores the response in a
// a JSON format, crawls,
// then graphs the crawl.
package main

import (
	"html/template"
	"net/http"
	"fmt"
)

type Crawl struct{
	Url string
	Keyword string
	Type string
	Crawl Type
}

type Type struct{
	B bool
	D bool
}

type Graph struct{
	//
}

func handler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("forms.html"))

	if r.Method != http.MethodPost {
		tmpl.Execute(w, nil)
		return
	}

	crawl := Crawl{
		Url: r.FormValue("Url"),
		Keyword: r.FormValue("Keyword"),
		Type: r.FormValue("Type"),
	}
	fmt.Println(crawl) // debug

	// TODO format crawl settings
	_ = crawl

	/* Crawler Program
			input: crawl in JSON
			output: graph in JSON */

	/* Render graph  */

	tmpl.Execute(w, struct{ Success bool }{true})
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":80", nil)
}
