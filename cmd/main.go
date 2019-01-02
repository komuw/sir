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

func main() {
	/*
		usage:
		  1. go run -race cmd/main.go
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
		time.AfterFunc(23*time.Second, clusterAndPlotRequests)
		// time.AfterFunc(23*time.Second, clusterAndPlotResponses)
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

type requestsResponses struct {
	l                      sync.RWMutex
	noOfAllRequests        int
	allRequests            []float64
	lengthOfLargestRequest int

	requestsSlice [][]byte

	noOfAllResponses     int
	allResponses         []float64
	lengthOfEachResponse int
}

var reqResp requestsResponses

func clusterAndPlotRequests() {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()

	for k, v := range reqResp.requestsSlice {
		diff := reqResp.lengthOfLargestRequest - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(nulByte), diff)
			v = append(v, pad...)
			reqResp.requestsSlice[k] = v
		}
	}
	for _, eachRequest := range reqResp.requestsSlice {
		for _, v := range eachRequest {
			reqResp.allRequests = append(reqResp.allRequests, float64(v))
		}
	}
	heart.Run(reqResp.noOfAllRequests, reqResp.lengthOfLargestRequest, reqResp.allRequests, 3.0, 1.0, false, "Requests")
}

func clusterAndPlotResponses() {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()
	heart.Run(reqResp.noOfAllResponses, reqResp.lengthOfEachResponse, reqResp.allResponses, 3.0, 1.0, false, "Responses")
}

func handleRequest(requestBuf []byte) {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()

	if reqResp.lengthOfLargestRequest < len(requestBuf) {
		reqResp.lengthOfLargestRequest = len(requestBuf)
	}

	// for _, v := range requestBuf {
	// 	reqResp.allRequests = append(reqResp.allRequests, float64(v))
	// }
	reqResp.requestsSlice = append(reqResp.requestsSlice, requestBuf)
}

func handleResponse(responseBuf []byte) {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()
	reqResp.lengthOfEachResponse = len(responseBuf)

	for _, v := range responseBuf {
		reqResp.allResponses = append(reqResp.allResponses, float64(v))
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

	requestBuf := new(bytes.Buffer)
	responseBuf := new(bytes.Buffer)
	ch := make(chan bool)

	// forward data from client to server
	go func() {
		tee := io.TeeReader(frontendConn, requestBuf)
		io.Copy(backendConn, tee)
		ch <- true
	}()

	// forward data from server to client
	go func() {
		tee := io.TeeReader(backendConn, responseBuf)
		io.Copy(frontendConn, tee)
		ch <- true
	}()

	<-ch
	<-ch
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
	requestBytes, err := ioutil.ReadAll(requestBuf)
	if err != nil {
		err = errors.Wrap(err, "unable to read & log request")
		log.Fatalf("%+v", err)
	}
	requestBytes = bytes.Trim(requestBytes, nulByte)
	handleRequest(requestBytes)
	log.Println("we sent request::", string(requestBytes))

	responseBytes, err := ioutil.ReadAll(responseBuf)
	if err != nil {
		err = errors.Wrap(err, "unable to read & log response")
		log.Fatalf("%+v", err)
	}
	handleResponse(responseBytes)
	log.Println("we got response::", string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////

	reqResp.l.Lock()
	reqResp.noOfAllRequests++
	reqResp.noOfAllResponses++
	log.Println("lengthOfLargestRequest:", reqResp.lengthOfLargestRequest)
	log.Println("lengthOfEachResponse:", reqResp.lengthOfEachResponse)
	reqResp.l.Unlock()
}
