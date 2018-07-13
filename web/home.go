// home.go runs the homepage,
// provides the user form,
// and stores the response
// in a JSON format.
package main

import (
	"html/template"
	"net/http"
	"fmt"
)

type FormPage struct{
	URL string
	Type string
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	p := FormPage{URL: "starting.url", Type: "Depth-first"} // Hardcoded values
	t, _ := template.ParseFiles("basictemplating.html")
	t.Execute(w, p)

}



func indexhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hey, %s</h1>", "<strong>there</strong>")
}

func main() {
	http.HandleFunc("/", indexhandler)
	http.HandleFunc("/form/", formHandler)
	http.ListenAndServe(":80", nil)
}
