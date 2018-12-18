package main

// 1. https://godoc.org/github.com/pa-m/sklearn/cluster
// 2. https://www.cheatography.com/chewxy/cheat-sheets/data-science-in-go-a/

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"

	"github.com/komuw/dbscan/proxyd"
	"github.com/pkg/errors"
)

/*
usage:
  go run -race reverse.go localhost:7777 localhost:3333
  curl -vkL localhost:7777
  echo -n "test out the server" | nc localhost 7777
*/

func forward(reverseProxyConn net.Conn) {
	defer reverseProxyConn.Close()
	err := reverseProxyConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to set reverseProxyConn deadline")
		log.Fatalf("\n%+v", err)
	}

	// TODO: make the buffer growable
	// TODO: use ioutil.ReadAll() for this
	buf := make([]byte, 96)
	reqLen, err := reverseProxyConn.Read(buf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Error reading")
		log.Fatalf("\n%+v", err)
	}
	_ = reqLen
	log.Println("Reverse read::", buf)
	log.Println("Reverse read2::", string(buf))

	// TODO: since we also want to dbscan the responses, we should make a copy here also.
	backendConn, err := net.Dial("tcp", os.Args[2])
	if err != nil {
		err = errors.Wrap(err, "Reverse Dial failed")
		log.Fatalf("\n%+v", err)
	}
	defer backendConn.Close()
	err = backendConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to set backendConn deadline")
		log.Fatalf("\n%+v", err)
	}
	log.Printf("reverseProxyConnected to localhost %v\n", reverseProxyConn)

	var backendBuf bytes.Buffer
	backendTee := io.TeeReader(backendConn, &backendBuf)
	io.Copy(backendConn, bytes.NewReader(buf))
	io.Copy(reverseProxyConn, backendTee)

	backendBytes, err := ioutil.ReadAll(&backendBuf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to read backendBuf")
		log.Fatalf("\n%+v", err)
	}
	log.Println("backendBytes:::", backendBytes)
	log.Println("backendBytes2:::", string(backendBytes))
}

func main() {
	{
		go proxyd.Run()
	}

	if len(os.Args) != 3 {
		log.Fatalf("\nReverse Usage %+v listen:port forward:port\n", os.Args[0])
		return
	}

	listener, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		err = errors.Wrapf(err, "Reverse failed to setup listener %v", os.Args[1])
		log.Fatalf("\n%+v", err)
	}

	for {
		reverseProxyConn, err := listener.Accept()
		if err != nil {
			err = errors.Wrapf(err, "Reverse failed to accept listener %v", os.Args[1])
			log.Fatalf("\n%+v", err)
		}
		log.Printf("Accepted reverseProxyConnection %v\n", reverseProxyConn)
		go forward(reverseProxyConn)
	}
}
