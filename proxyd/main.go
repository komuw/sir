package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

/*
usage:
  export SERVERNAME=candidate && go run -race proxyd/main.go
  curl -vL \
	-H "accept: application/json" \
	-H "Content-Type: application/json" \
	-H "Host: httpbin.org" \
	-d '{"name":"longerRequest"}' \
	localhost:3000/post

build:
  export GOOS=linux && \
  export GOARCH=amd64 && \
  export CGO_ENABLED=0 && \
  go build -ldflags="-s -w" -o proxyd/mockproxy proxyd/main.go
*/
var norequets uint16

func main() {
	http.HandleFunc("/post", echoHandler)

	log.Println(`
	listening at localhost:3000
	send a request like:
	  curl -vL \
		-H "accept: application/json" \
		-H "Content-Type: application/json" \
		-H "Host: httpbin.org" \
		-d '{"name":"longerRequest"}' \
		localhost:3000/post`)
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		err = errors.Wrap(err, "unable to listen and serve on port 3000")
		log.Fatalf("%+v", err)
	}
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = errors.Wrap(err, "unable to read request body")
		log.Printf("%+v", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Connection", "close") // important so that clients do not have to wait
	w.Header().Set("Content-Type", "application/json")
	w.Write(reqBytes)

	norequets++
	serverName, _ := os.LookupEnv("SERVERNAME")
	log.Printf("%v succesfully handled request number %v in %v secs", serverName, norequets, time.Since(start).Seconds())
}
