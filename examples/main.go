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

type requestsResponses struct {
	l                   sync.RWMutex
	noOfAllRequests     int
	allRequests         []float64
	lengthOfEachRequest int

	noOfAllResponses     int
	allResponses         []float64
	lengthOfEachResponse int
}

var reqResp requestsResponses

func clusterAndPlotRequests() {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()
	heart.Run(reqResp.noOfAllRequests, reqResp.lengthOfEachRequest, reqResp.allRequests, 3.0, 1.0, false)
}

func clusterAndPlotResponses() {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()
	heart.Run(reqResp.noOfAllResponses, reqResp.lengthOfEachResponse, reqResp.allResponses, 3.0, 1.0, false)
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
		time.AfterFunc(23*time.Second, clusterAndPlotRequests)
		time.AfterFunc(23*time.Second, clusterAndPlotResponses)
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

func handleRequest(requestBuf []byte) {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()
	reqResp.lengthOfEachRequest = len(requestBuf)

	for _, v := range requestBuf {
		reqResp.allRequests = append(reqResp.allRequests, float64(v))
	}
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

	//////////////////////////////////// LOG REQUEST ////////////////////////
	// TODO: make the buffer growable
	requestBuf := make([]byte, 1024)
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
	handleResponse(backendBytes)
	log.Println("we got response::", backendBytes)
	log.Println("we got response::", string(backendBytes))
	//////////////////////////////////// LOG RESPONSE ////////////////////////

	reqResp.l.Lock()
	reqResp.noOfAllRequests++
	reqResp.noOfAllResponses++
	log.Println("allRequests:", reqResp.allRequests)
	log.Println("allResponses:", reqResp.allResponses)
	log.Println("lengthOfEachRequest:", reqResp.lengthOfEachRequest)
	log.Println("lengthOfEachResponse:", reqResp.lengthOfEachResponse)
	reqResp.l.Unlock()
}
