# Graphical Web Crawler 
Welcome!

## Team Crater 
* Helen Stockman
* James Wilson
* Dominic Phan

## Instructions
Skip to your system build if you already have Go installed. 

### [Download and Install Go][install]
Work through [How to Write Go Code][code] to set up your workspace. 

[install]: https://golang.org/doc/install
[code]: https://golang.org/doc/code.html

### Build and run

#### Linux 

##### Quick
```
$ go get github.com/helenuria/go-crawler
$ [sudo] $GOPATH/bin/web
```

##### Development
```
$ go get github.com/helenuria/go-crawler
$ cd $GOPATH/src/github.com/helenuria/go-crawler/web
```
Make some changes or build as is.
```
$ go build
$ [sudo] ./web
```

#### Windows

##### Quick
```
$ go get golang.org/x/net/html
$ go get github.com/helenuria/go-crawler
$ $GOPATH/bin/web.exe
```

##### Development 
```
$ go get golang.org/x/net/html
$ go get github.com/helenuria/go-crawler
$ cd $GOPATH/src/github.com/helenuria/go-crawler/web
```
Make some changes or build as is.
```
$ go build
$ web.exe
```

#### Mac build

##### Quick 
```
$ go get github.com/helenuria/go-crawler
$ ./$GOPATH/bin/web -addr :8080
```

##### Development
```
$ go get github.com/helenuria/go-crawler
$ cd $GOPATH/src/github.com/helenuria/go-crawler/web
```
Make some changes or build as is.
```
$ go build
$ ./web -addr :8080
```

### Usage 
* Navigate to `localhost:[port]` in your browser. Where port is `80` or `8080` depending on your build. 
* Submit some data and watch the console output how the submitted form data is stored.
* Interactive graph displayed after posting form.


