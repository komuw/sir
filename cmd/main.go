package main

import (
	"bytes"
	"fmt"
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
	primaryBackendAddr := "google.com:80"
	secondaryBackendAddr := "google.com:80" //"bing.com:80"

	reqRespCandidate := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Candidate, Addr: candidateBackendAddr}}
	reqRespPrimary := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Primary, Addr: primaryBackendAddr}}
	reqRespSecondary := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Secondary, Addr: secondaryBackendAddr}}

	{
		// candidate
		clusterAndPlotReqCandidate := func() {
			reqRespCandidate.ClusterAndPlotRequests()
		}
		clusterAndPlotResCandidate := func() {
			reqRespCandidate.ClusterAndPlotResponses()
		}
		time.AfterFunc(23*time.Second, clusterAndPlotReqCandidate)
		time.AfterFunc(23*time.Second, clusterAndPlotResCandidate)

		// primary
		clusterAndPlotReqPrimary := func() {
			reqRespPrimary.ClusterAndPlotRequests()
		}
		clusterAndPlotResPrimary := func() {
			reqRespPrimary.ClusterAndPlotResponses()
		}
		time.AfterFunc(25*time.Second, clusterAndPlotReqPrimary)
		time.AfterFunc(25*time.Second, clusterAndPlotResPrimary)

		// secondary
		clusterAndPlotReqSecondary := func() {
			reqRespSecondary.ClusterAndPlotRequests()
		}
		clusterAndPlotResSecondary := func() {
			reqRespSecondary.ClusterAndPlotResponses()
		}
		time.AfterFunc(27*time.Second, clusterAndPlotReqSecondary)
		time.AfterFunc(27*time.Second, clusterAndPlotResSecondary)

		//TODO:
		//1. this time.AfterFuncs should all be scheduled to run at the same time
		//2. actually, we should not be using time.AfterFunc at all; but some other mechanism
	}

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
		var rb = make(chan []byte)
		go forward(frontendConn, reqRespCandidate, rb)
		request := <-rb
		go priSecForward(request, reqRespPrimary)
		time.Sleep(3 * time.Second) // TODO: remove this sleeps
		go priSecForward(request, reqRespSecondary)

		fmt.Println()
		fmt.Println("request", request, string(request))
		fmt.Println()
	}
}

func forward(frontendConn net.Conn, reqResp *sir.RequestsResponse, rb chan []byte) {
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "unable to set reverseProxyConn deadline")
		log.Fatalf("%+v", err)
	}

	backendConn, err := net.Dial("tcp", reqResp.Backend.Addr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrapf(err, "unable to set backendConn deadline of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	log.Printf("frontend connected to backend %v", reqResp.Backend)

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

	requestBytes, err := ioutil.ReadAll(requestBuf)
	if err != nil {
		err = errors.Wrap(err, "unable to read & log request")
		log.Fatalf("%+v", err)
	}
	requestBytes = bytes.Trim(requestBytes, sir.NulByte)
	reqResp.HandleRequest(requestBytes)
	rb <- requestBytes

	responseBytes, err := ioutil.ReadAll(responseBuf)
	if err != nil {
		err = errors.Wrapf(err, "unable to read & log response of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleResponse(responseBytes)

	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
	log.Printf("we sent request to backend %v \n %v", reqResp.Backend, string(requestBytes))
	log.Printf("we got response from backend %v \n %v", reqResp.Backend, string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
}

func priSecForward(requestBytes []byte, reqResp *sir.RequestsResponse) {
	backendConn, err := net.Dial("tcp", reqResp.Backend.Addr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrapf(err, "unable to set backendConn deadline of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	log.Printf("frontend connected to backend %v", reqResp.Backend)

	_, err = backendConn.Write(requestBytes)
	if err != nil {
		err = errors.Wrapf(err, "backendConn.Write failed for backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleRequest(requestBytes)

	responseBytes, err := ioutil.ReadAll(backendConn)
	if err != nil {
		err = errors.Wrapf(err, "unable to read & log response of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleResponse(responseBytes)

	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
	log.Printf("we sent request to backend %v \n %v", reqResp.Backend, string(requestBytes))
	log.Printf("we got response from backend %v \n %v", reqResp.Backend, string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
}
