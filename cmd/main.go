package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"github.com/komuw/sir/pkg"
	"github.com/pkg/errors"
)

func main() {
	/*
		usage:
		  1. go run -race cmd/main.go
		  2. curl -vL -H "Host: httpbin.org" localhost:7777/get
	*/
	frontendAddr := "localhost:7777"
	candidateBackendAddr := "httpbin.org:80"
	primaryBackendAddr := "httpbin.org:80"
	secondaryBackendAddr := "httpbin.org:80"

	listener, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		log.Fatalf("failed to setup listener %v", err)
	}
	log.Println("Sir Listening on " + frontendAddr)
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
		log.Printf("ready to accept connections to frontend %v", frontendAddr)

		// TODO: remove the sleeps
		go forward(frontendConn, candidateBackendAddr, candidate)
		time.Sleep(2 * time.Second)
		go forward(frontendConn, primaryBackendAddr, primary)
		time.Sleep(2 * time.Second)
		go forward(frontendConn, secondaryBackendAddr, secondary)
	}
}

type requestsResponses struct {
	l                      sync.RWMutex
	noOfAllRequests        int
	allRequests            []float64
	lengthOfLargestRequest int
	requestsSlice          [][]byte

	noOfAllResponses        int
	allResponses            []float64
	lengthOfLargestResponse int
	responsesSlice          [][]byte
}

var reqResp requestsResponses

func clusterAndPlotRequests() {
	appendName := "Requests"
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
	nclusters, X, err := sir.GetClusters(reqResp.noOfAllRequests, reqResp.lengthOfLargestRequest, reqResp.allRequests, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Requests estimated number of clusters: %d\n", nclusters)

	proj := sir.FindPCA(X, reqResp.lengthOfLargestRequest)
	err = sir.PlotResultsPCA(reqResp.noOfAllRequests, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

func clusterAndPlotResponses() {
	appendName := "Responses"
	reqResp.l.Lock()
	defer reqResp.l.Unlock()

	for k, v := range reqResp.responsesSlice {
		diff := reqResp.lengthOfLargestResponse - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(nulByte), diff)
			v = append(v, pad...)
			reqResp.responsesSlice[k] = v
		}
	}
	for _, eachResponse := range reqResp.responsesSlice {
		for _, v := range eachResponse {
			reqResp.allResponses = append(reqResp.allResponses, float64(v))
		}
	}
	nclusters, X, err := sir.GetClusters(reqResp.noOfAllResponses, reqResp.lengthOfLargestResponse, reqResp.allResponses, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Responses stimated number of clusters: %d\n", nclusters)

	proj := sir.FindPCA(X, reqResp.lengthOfLargestResponse)
	err = sir.PlotResultsPCA(reqResp.noOfAllResponses, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

func handleRequest(requestBuf []byte) {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()

	if reqResp.lengthOfLargestRequest < len(requestBuf) {
		reqResp.lengthOfLargestRequest = len(requestBuf)
	}
	reqResp.requestsSlice = append(reqResp.requestsSlice, requestBuf)
}

func handleResponse(responseBuf []byte) {
	reqResp.l.Lock()
	defer reqResp.l.Unlock()

	if reqResp.lengthOfLargestResponse < len(responseBuf) {
		reqResp.lengthOfLargestResponse = len(responseBuf)
	}
	reqResp.responsesSlice = append(reqResp.responsesSlice, responseBuf)
}

const nulByte = "\x00"

type backendType int

const (
	candidate backendType = iota
	primary
	secondary
)

func (backend backendType) String() string {
	names := []string{
		"candidate",
		"primary",
		"secondary"}
	return names[backend]
}

func forward(frontendConn net.Conn, remoteAddr string, backend backendType) {
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "unable to set reverseProxyConn deadline")
		log.Fatalf("%+v", err)
	}

	backendConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for address %s of backend %v", remoteAddr, backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrapf(err, "unable to set backendConn deadline of backend %v", backend)
		log.Fatalf("%+v", err)
	}
	log.Printf("frontend connected to backend %v(%v)", backend, remoteAddr)

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
	log.Printf("we sent request to backend %v \n %v", backend, string(requestBytes))

	responseBytes, err := ioutil.ReadAll(responseBuf)
	if err != nil {
		err = errors.Wrapf(err, "unable to read & log response of backend %v", backend)
		log.Fatalf("%+v", err)
	}
	handleResponse(responseBytes)
	log.Printf("we got response from backend %v \n %v", backend, string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////

	reqResp.l.Lock()
	reqResp.noOfAllRequests++
	reqResp.noOfAllResponses++
	log.Printf("lengthOfLargestRequest for backend %v %v", backend, reqResp.lengthOfLargestRequest)
	log.Printf("lengthOfLargestResponse for backend %v %v", backend, reqResp.lengthOfLargestResponse)
	reqResp.l.Unlock()
}
