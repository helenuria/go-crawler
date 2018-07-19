# Graphical Web Crawler #
Welcome!

## Team Crater ##
* Helen Stockman
* James Wilson
* Dominic Phan

### Windows build ###
* navigate to the project's "/web/" dir
* run the command "go get golang.org/x/net/html"
* run the command "go build"
* run "web.exe"
* navigate to localhost:80 in your browser.
* Submit some data and watch the console output how the submitted form data is stored.
* Interactive crawl graph at localhost/example/

### Mac build ###
* navigate to the project's `/web` directory
* run the command `go build`
* run the command `./web -addr :8080`
* navigate to `localhost:8080` in your browser.
* Submit some data and watch the console output how the submitted form data is stored.
* Interactive crawl graph at localhost/example/

#### home.go ####
* web server
* listens on port 79
* serves forms.html to the client/browser
* collects the form input and stores it in the 'crawl' struct
* handles the crawler program and graph rendering
