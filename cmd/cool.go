package main

import (
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/komuw/sir/pkg"
	"github.com/pkg/errors"
)

func priSecForward(requestBytes []byte, remoteAddr string, reqResp *sir.RequestsResponse) {
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

	_, err = backendConn.Write(requestBytes)
	if err != nil {
		err = errors.Wrapf(err, "backendConn.Write of backend %v failrd", reqResp.Backend)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleRequest(requestBytes)

	responseBytes, err := ioutil.ReadAll(backendConn)
	if err != nil {
		err = errors.Wrapf(err, "unable to read & log response of backend %v(%v)", reqResp.Backend, remoteAddr)
		log.Fatalf("%+v", err)
	}
	reqResp.HandleResponse(responseBytes)

	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////
	log.Printf("we sent request to backend %v(%v) \n %v", reqResp.Backend, remoteAddr, string(requestBytes))
	log.Printf("we got response from backend %v(%v) \n %v", reqResp.Backend, remoteAddr, string(responseBytes))
	//////////////////////////////////// LOG REQUEST  & RESPONSE ////////////////////////

	reqResp.L.Lock()
	reqResp.NoOfAllRequests++
	reqResp.NoOfAllResponses++
	log.Printf("lengthOfLargestRequest for backend %v(%v) %v", reqResp.Backend, remoteAddr, reqResp.LengthOfLargestRequest)
	log.Printf("lengthOfLargestResponse for backend %v(%v) %v", reqResp.Backend, remoteAddr, reqResp.LengthOfLargestResponse)
	reqResp.L.Unlock()
}
