package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
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

	// secondaryBackendAddr := "httpbin.org:80"

	reqRespCandidate := &sir.RequestsResponse{Backend: sir.Candidate}
	// reqRespSecondary := &sir.RequestsResponse{Backend: sir.Secondary}
	// {
	// 	// candidate
	// 	clusterAndPlotReqCandidate := func() {
	// 		reqRespCandidate.ClusterAndPlotRequests()
	// 	}
	// 	clusterAndPlotResCandidate := func() {
	// 		reqRespCandidate.ClusterAndPlotResponses()
	// 	}
	// 	time.AfterFunc(23*time.Second, clusterAndPlotReqCandidate)
	// 	time.AfterFunc(23*time.Second, clusterAndPlotResCandidate)

	// 	// primary
	// 	clusterAndPlotReqPrimary := func() {
	// 		reqRespPrimary.ClusterAndPlotRequests()
	// 	}
	// 	clusterAndPlotResPrimary := func() {
	// 		reqRespPrimary.ClusterAndPlotResponses()
	// 	}
	// 	time.AfterFunc(25*time.Second, clusterAndPlotReqPrimary)
	// 	time.AfterFunc(25*time.Second, clusterAndPlotResPrimary)

	// 	// secondary
	// 	clusterAndPlotReqSecondary := func() {
	// 		reqRespSecondary.ClusterAndPlotRequests()
	// 	}
	// 	clusterAndPlotResSecondary := func() {
	// 		reqRespSecondary.ClusterAndPlotResponses()
	// 	}
	// 	time.AfterFunc(27*time.Second, clusterAndPlotReqSecondary)
	// 	time.AfterFunc(27*time.Second, clusterAndPlotResSecondary)

	// 	//TODO:
	// 	//1. this time.AfterFuncs should all be scheduled to run at the same time
	// 	//2. actually, we should not be using time.AfterFunc at all; but some other mechanism
	// }

	listener, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		log.Fatalf("failed to setup listener %v", err)
	}
	log.Println("Sir Listening on " + frontendAddr)
	log.Println(`
	To use it, send a request like:
	    curl -vL -H "Host: httpbin.org" localhost:7777/get
	`)

	for {
		frontendConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept listener %v", err)
		}
		log.Printf("ready to accept connections to frontend %v", frontendAddr)

		// TODO: remove the sleeps
		go forward(frontendConn, candidateBackendAddr, reqRespCandidate)
		time.Sleep(2 * time.Second)
		// time.Sleep(2 * time.Second)
		// go forward(frontendConn, secondaryBackendAddr, reqRespSecondary)
	}
}

var primaryBackendAddr = "google.com:80"
var reqRespPrimary = &sir.RequestsResponse{Backend: sir.Primary}

func forward(frontendConn net.Conn, remoteAddr string, reqResp *sir.RequestsResponse) {
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "unable to set reverseProxyConn deadline")
		log.Fatalf("%+v", err)
	}

	backendConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for address %s of backend %v", remoteAddr, reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrapf(err, "unable to set backendConn deadline of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	log.Printf("frontend connected to backend %v(%v)", reqResp.Backend, remoteAddr)

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
	requestBytes = bytes.Trim(requestBytes, sir.NulByte)
	reqResp.HandleRequest(requestBytes)
	log.Printf("we sent request to backend %v \n %v", reqResp.Backend, string(requestBytes))

	responseBytes, err := ioutil.ReadAll(responseBuf)
	if err != nil {
		err = errors.Wrapf(err, "unable to read & log response of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleResponse(responseBytes)
	log.Printf("we got response from backend %v \n %v", reqResp.Backend, string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////

	reqResp.L.Lock()
	reqResp.NoOfAllRequests++
	reqResp.NoOfAllResponses++
	log.Printf("lengthOfLargestRequest for backend %v %v", reqResp.Backend, reqResp.LengthOfLargestRequest)
	log.Printf("lengthOfLargestResponse for backend %v %v", reqResp.Backend, reqResp.LengthOfLargestResponse)
	reqResp.L.Unlock()

	time.Sleep(2 * time.Second)
	go primaryForward(requestBytes, primaryBackendAddr, reqRespPrimary)
}
