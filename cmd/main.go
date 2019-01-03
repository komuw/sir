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

	reqRespCandidate := &sir.RequestsResponses{Backend: sir.Candidate}
	reqRespPrimary := &sir.RequestsResponses{Backend: sir.Primary}
	reqRespSecondary := &sir.RequestsResponses{Backend: sir.Secondary}
	{
		// candidate
		clusterAndPlotReqCandidate := func() {
			clusterAndPlotRequests(reqRespCandidate)
		}
		clusterAndPlotResCandidate := func() {
			clusterAndPlotResponses(reqRespCandidate)
		}
		time.AfterFunc(23*time.Second, clusterAndPlotReqCandidate)
		time.AfterFunc(23*time.Second, clusterAndPlotResCandidate)

		// primary
		clusterAndPlotReqPrimary := func() {
			clusterAndPlotRequests(reqRespPrimary)
		}
		clusterAndPlotResPrimary := func() {
			clusterAndPlotResponses(reqRespPrimary)
		}
		time.AfterFunc(23*time.Second, clusterAndPlotReqPrimary)
		time.AfterFunc(23*time.Second, clusterAndPlotResPrimary)

		// secondary
		clusterAndPlotReqSecondary := func() {
			clusterAndPlotRequests(reqRespSecondary)
		}
		clusterAndPlotResSecondary := func() {
			clusterAndPlotResponses(reqRespSecondary)
		}
		time.AfterFunc(23*time.Second, clusterAndPlotReqSecondary)
		time.AfterFunc(23*time.Second, clusterAndPlotResSecondary)
	}

	for {
		frontendConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept listener %v", err)
		}
		log.Printf("ready to accept connections to frontend %v", frontendAddr)

		// TODO: remove the sleeps
		go forward(frontendConn, candidateBackendAddr, reqRespCandidate)
		time.Sleep(2 * time.Second)
		go forward(frontendConn, primaryBackendAddr, reqRespPrimary)
		time.Sleep(2 * time.Second)
		go forward(frontendConn, secondaryBackendAddr, reqRespSecondary)
	}
}

func clusterAndPlotRequests(reqResp *sir.RequestsResponses) {
	appendName := "Requests"
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	for k, v := range reqResp.RequestsSlice {
		diff := reqResp.LengthOfLargestRequest - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(sir.NulByte), diff)
			v = append(v, pad...)
			reqResp.RequestsSlice[k] = v
		}
	}
	for _, eachRequest := range reqResp.RequestsSlice {
		for _, v := range eachRequest {
			reqResp.AllRequests = append(reqResp.AllRequests, float64(v))
		}
	}
	nclusters, X, err := sir.GetClusters(reqResp.NoOfAllRequests, reqResp.LengthOfLargestRequest, reqResp.AllRequests, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Requests estimated number of clusters: %d\n", nclusters)

	proj := sir.FindPCA(X, reqResp.LengthOfLargestRequest)
	err = sir.PlotResultsPCA(reqResp.NoOfAllRequests, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

func clusterAndPlotResponses(reqResp *sir.RequestsResponses) {
	appendName := "Responses"
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	for k, v := range reqResp.ResponsesSlice {
		diff := reqResp.LengthOfLargestResponse - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(sir.NulByte), diff)
			v = append(v, pad...)
			reqResp.ResponsesSlice[k] = v
		}
	}
	for _, eachResponse := range reqResp.ResponsesSlice {
		for _, v := range eachResponse {
			reqResp.AllResponses = append(reqResp.AllResponses, float64(v))
		}
	}
	nclusters, X, err := sir.GetClusters(reqResp.NoOfAllResponses, reqResp.LengthOfLargestResponse, reqResp.AllResponses, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Responses stimated number of clusters: %d\n", nclusters)

	proj := sir.FindPCA(X, reqResp.LengthOfLargestResponse)
	err = sir.PlotResultsPCA(reqResp.NoOfAllResponses, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

func forward(frontendConn net.Conn, remoteAddr string, reqResp *sir.RequestsResponses) {
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
}
