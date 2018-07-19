# Graphical Web Crawler
Welcome!

## Team Crater
* Helen Stockman
* James Wilson
* Dominic Phan


## home.go 

### home.go is a web server that:
* listens on port 80
* serves forms.html to the client/browser
* collects the form input and stores it in the 'crawl' struct
* is structured to handle the crawler program and graph rendering

To test (in windows)
* navigate to the project's "/web/" dir 
* run the command "go build"
* run "web.exe"
* navigate to localhost:80 in your browser. 
* Submit some data and watch the console output how the submitted form data is stored.

To test (on mac)
* navigate to the project's `/web` directory
* run the command `go build`
* run the command `./web -addr :8080`
* navigate to `localhost:8080` in your browser. 
* Submit some data and watch the console output how the submitted form data is stored.
