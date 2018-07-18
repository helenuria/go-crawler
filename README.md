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

To test, navigate to the project's "/web/" dir and run the command, "go build", in the console to build, "web.exe". Run "web.exe" and navigate to localhost:80 in your browser. Submit some data and watch the console output how the submitted form data is stored.
