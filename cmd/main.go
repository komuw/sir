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

// TODO: make this configurable
const netTimeouts = 6 * time.Second
const thresholdOfClusterCalculation = 100

func main() {
	/*
		usage:
		  1. go run -race cmd/main.go
		  2. curl -vL -H "Host: httpbin.org" localhost:7777/get
	*/
	frontendAddr := "localhost:7777"
	candidateBackendAddr := "localhost:3001" //"httpbin.org:80"
	primaryBackendAddr := "localhost:3002"   //"google.com:80"
	secondaryBackendAddr := "localhost:3003" //"bing.com:80"

	reqRespCandidate := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Candidate, Addr: candidateBackendAddr}}
	reqRespPrimary := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Primary, Addr: primaryBackendAddr}}
	reqRespSecondary := &sir.RequestsResponse{Backend: sir.Backend{Type: sir.Secondary, Addr: secondaryBackendAddr}}

	listener, err := net.Listen("tcp", frontendAddr)
	if err != nil {
		err = errors.Wrapf(err, "failed to setup listener for address %v", frontendAddr)
		log.Fatalf("%+v", err)
	}
	log.Printf(`
	Sir is running and listening on %v
	To use it, send a request like:
	    curl -vL -H "Host: httpbin.org" localhost:7777/get
	`, frontendAddr)

	for {
		frontendConn, err := listener.Accept()
		if err != nil {
			err = errors.Wrapf(err, "failed to accept listener for address %v", frontendAddr)
			log.Fatalf("%+v", err)
		}
		log.Printf("ready to accept connections to frontend %v", frontendAddr)

		if calculateThreshold(reqRespCandidate.NoOfAllRequests, thresholdOfClusterCalculation) {
			go clusterPlot(reqRespPrimary, reqRespCandidate)
			// go clusterPlot(reqRespPrimary, thresholdOfClusterCalculation)
			// go clusterPlot(reqRespSecondary, thresholdOfClusterCalculation)

			resetC := &sir.RequestsResponse{
				Backend: sir.Backend{Type: reqRespCandidate.Backend.Type, Addr: reqRespCandidate.Backend.Addr}}
			reqRespCandidate = resetC

			resetP := &sir.RequestsResponse{
				Backend: sir.Backend{Type: reqRespPrimary.Backend.Type, Addr: reqRespPrimary.Backend.Addr}}
			reqRespPrimary = resetP

			resetS := &sir.RequestsResponse{
				Backend: sir.Backend{Type: reqRespSecondary.Backend.Type, Addr: reqRespSecondary.Backend.Addr}}
			reqRespSecondary = resetS
		}

		var rb = make(chan []byte)
		go forward(frontendConn, reqRespCandidate, rb)
		request := <-rb
		go priSecForward(request, reqRespPrimary)
		go priSecForward(request, reqRespSecondary)
	}
}

func calculateThreshold(noOfRequests, threshold int) bool {
	if noOfRequests == 0 {
		noOfRequests = 1
	}
	return (noOfRequests % threshold) == 0
}

func clusterPlot(major *sir.RequestsResponse, minor *sir.RequestsResponse) {
	major.L.Lock()
	defer major.L.Unlock()
	minor.L.Lock()
	defer minor.L.Unlock()

	sir.ClusterAndPlotRequests(major, minor)
	// reqResp.ClusterAndPlotRequests()
	// reqResp.ClusterAndPlotResponses()
}

func forward(frontendConn net.Conn, reqResp *sir.RequestsResponse, rb chan []byte) {
	start := time.Now()
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(netTimeouts))
	if err != nil {
		err = errors.Wrap(err, "unable to set frontendConn deadline")
		log.Fatalf("%+v", err)
	}

	dialer := net.Dialer{Timeout: netTimeouts, DualStack: true, FallbackDelay: 20 * time.Millisecond}
	backendConn, err := dialer.Dial("tcp", reqResp.Backend.Addr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(netTimeouts))
	if err != nil {
		err = errors.Wrapf(err, "unable to set backendConn deadline of backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	log.Printf("frontend connected to backend %v", reqResp.Backend)

	requestBuf := new(bytes.Buffer)
	responseBuf := new(bytes.Buffer)
	ch := make(chan struct{}, 2)

	// forward data from client to server
	go func() {
		tee := io.TeeReader(frontendConn, requestBuf)
		io.Copy(backendConn, tee)
		ch <- struct{}{}
	}()

	// forward data from server to client
	go func() {
		tee := io.TeeReader(backendConn, responseBuf)
		io.Copy(frontendConn, tee)
		ch <- struct{}{}
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
	log.Printf("we got response from backend %v in %v secs \n %v", reqResp.Backend, time.Since(start).Seconds(), string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
}

func priSecForward(requestBytes []byte, reqResp *sir.RequestsResponse) {
	start := time.Now()
	dialer := net.Dialer{Timeout: netTimeouts, DualStack: true, FallbackDelay: 20 * time.Millisecond}
	backendConn, err := dialer.Dial("tcp", reqResp.Backend.Addr)
	if err != nil {
		err = errors.Wrapf(err, "dial failed for backend %v", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(netTimeouts))
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
	log.Printf("we got response from backend %v in %v secs \n %v", reqResp.Backend, time.Since(start).Seconds(), string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////

}
