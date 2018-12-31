package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/komuw/dbscan/heart"
	"github.com/pkg/errors"
)

func cooler() {
	thisStruct.l.Lock()
	defer thisStruct.l.Unlock()
	heart.Run(thisStruct.noOfAllRequests, thisStruct.lengthOfEachRequest, thisStruct.allRequests, 3.0, 1.0, false)
}

func main() {
	/*
		usage:
		  1. go run -race main.go
		  2. curl -vL -H "Host: httpbin.org" localhost:7777/get
	*/
	frontendAddr := "localhost:7777"
	backendAddr := "httpbin.org:80" // why is it that "httpbin.org:443" does not work

	listener, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		log.Fatalf("failed to setup listener %v", err)
	}
	log.Println("ReverseProxy Listening on " + frontendAddr)
	log.Println(`
	To use it, send a request like:
	    curl -vL -H "Host: httpbin.org" localhost:7777/get
	`)

	{
		time.AfterFunc(23*time.Second, cooler)
	}

	for {
		frontendConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept listener %v", err)
		}
		log.Print("Accepted frontendConn")

		go forward(frontendConn, backendAddr)
	}
}

type myStruct struct {
	l                   sync.RWMutex
	noOfAllRequests     int
	allRequests         []float64
	lengthOfEachRequest int
}

var thisStruct myStruct

func handleRequest(requestBuf []byte) {
	thisStruct.l.Lock()
	defer thisStruct.l.Unlock()
	thisStruct.lengthOfEachRequest = len(requestBuf)

	for _, v := range requestBuf {
		thisStruct.allRequests = append(thisStruct.allRequests, float64(v))
	}
}

const nulByte = "\x00"

func forward(frontendConn net.Conn, remoteAddr string) {
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "unable to set reverseProxyConn deadline")
		log.Fatalf("%+v", err)
	}

	//////////////////////////////////// LOG REQUEST ////////////////////////
	// TODO: make the buffer growable
	requestBuf := make([]byte, 1024, 1024)
	reqLen, err := frontendConn.Read(requestBuf)
	if err != nil {
		log.Fatalf("Error reading %+v", err)
	}
	_ = reqLen
	requestBuf = bytes.Trim(requestBuf, nulByte)
	handleRequest(requestBuf)
	log.Println("we sent request::", requestBuf)
	log.Println("we sent request::", string(requestBuf))
	//////////////////////////////////// LOG REQUEST ////////////////////////

	backendConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		err = errors.Wrap(err, "dial failed for address"+remoteAddr)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "unable to set backendConn deadline")
		log.Fatalf("%+v", err)
	}
	log.Print("frontendConnected")

	var backendBuf bytes.Buffer
	backendTee := io.TeeReader(backendConn, &backendBuf)
	io.Copy(backendConn, bytes.NewReader(requestBuf))
	io.Copy(frontendConn, backendTee)

	//////////////////////////////////// LOG RESPONSE ////////////////////////
	backendBytes, err := ioutil.ReadAll(&backendBuf)
	if err != nil {
		err = errors.Wrap(err, "unable to read backendBuf")
		log.Fatalf("%+v", err)
	}
	log.Println("we got response::", backendBytes)
	log.Println("we got response::", string(backendBytes))
	//////////////////////////////////// LOG RESPONSE ////////////////////////

	thisStruct.l.Lock()
	thisStruct.noOfAllRequests++
	log.Println("allRequests:", thisStruct.allRequests)
	log.Println("lengthOfEachRequest:", thisStruct.lengthOfEachRequest)
	thisStruct.l.Unlock()
}

// Hello.
// I'm trying to create a TCP reverse proxy.
// It should work for both http/https/rabbitMQ etc, anything that talks tcp.

// The way it should work is:
// 1. A client sends requests to it.
// 2. The reverse proxy logs the resquest(by reading it)
// 3. The reverse proxy forwards the request to a backend
// 4. When the backend responds, the reverse proxy logs the response(by reading it)
// 5. The reverse proxy forwards the response to the client

// I have two problems/questions.
// 1. In the code to log request
// `requestBuf := make([]byte, 512)`
// how do I make the buffer growable?
// I want to make it possible to handle a request of arbitrary size.

// 2. Why does my reverse proxy fail to work for https?
// ie, when I set
// `backendAddr := "httpbin.org:443"`
// the reverse proxy does not work anymore.

// Thanks
