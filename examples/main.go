package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"
)

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

	for {
		frontendConn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept listener %v", err)
		}
		log.Print("Accepted frontendConn")

		go forward(frontendConn, backendAddr)
	}
}

func forward(frontendConn net.Conn, remoteAddr string) {
	defer frontendConn.Close()
	err := frontendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Fatalf("Unable to set frontendConn deadline %v", err)
	}

	// TODO: make the buffer growable
	requestBuf := make([]byte, 512)
	reqLen, err := frontendConn.Read(requestBuf)
	if err != nil {
		log.Fatalf("Error reading %+v", err)
	}
	_ = reqLen
	log.Println("we sent request::", requestBuf)
	log.Println("we sent request::", string(requestBuf))

	backendConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		msg := "Dial failed for address" + remoteAddr
		log.Fatalf("%+v %+v", msg, err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Fatalf("Unable to set backendConn deadline %v", err)
	}
	log.Print("frontendConnected")

	var backendBuf bytes.Buffer
	backendTee := io.TeeReader(backendConn, &backendBuf)
	io.Copy(backendConn, bytes.NewReader(requestBuf))
	io.Copy(frontendConn, backendTee)

	backendBytes, err := ioutil.ReadAll(&backendBuf)
	if err != nil {

		log.Fatalf("Unable to read backendBuf %+v", err)
	}
	log.Println("we got response::", backendBytes)
	log.Println("we got response::", string(backendBytes))
}
