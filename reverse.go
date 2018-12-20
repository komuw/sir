package main

// 1. https://godoc.org/github.com/pa-m/sklearn/cluster
// 2. https://www.cheatography.com/chewxy/cheat-sheets/data-science-in-go-a/

import (
	"bytes"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/komuw/dbscan/proxyd"
	"github.com/pkg/errors"
)

/*
usage:
  go run -race reverse.go -p localhost:7777 -r localhost:3333
  curl -vkL localhost:7777
  echo -n "test out the server" | nc localhost 7777
*/

// TODO: do the same for responses
var noOfAllRequests = 0

var allRequests []float64
var lengthOfEachRequest = 0

func handleRequest(requestBuf []byte) {
	lengthOfEachRequest = len(requestBuf)

	for _, v := range requestBuf {
		allRequests = append(allRequests, float64(v))
	}
}

func forward(reverseProxyConn net.Conn, remoteAddr string) {
	defer reverseProxyConn.Close()
	err := reverseProxyConn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to set reverseProxyConn deadline")
		log.Fatalf("\n%+v", err)
	}

	// TODO: make the buffer growable
	// TODO: use ioutil.ReadAll() for this
	requestBuf := make([]byte, 96)
	reqLen, err := reverseProxyConn.Read(requestBuf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Error reading")
		log.Fatalf("\n%+v", err)
	}
	_ = reqLen
	log.Println("Reverse read::", requestBuf)
	log.Println("Reverse read2::", string(requestBuf))
	handleRequest(requestBuf)

	// TODO: since we also want to dbscan the responses, we should make a copy here also.
	backendConn, err := net.Dial("tcp", remoteAddr)
	if err != nil {
		err = errors.Wrap(err, "Reverse Dial failed for address"+remoteAddr)
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
	io.Copy(backendConn, bytes.NewReader(requestBuf))
	io.Copy(reverseProxyConn, backendTee)

	backendBytes, err := ioutil.ReadAll(&backendBuf)
	if err != nil {
		err = errors.Wrap(err, "Reverse Unable to read backendBuf")
		log.Fatalf("\n%+v", err)
	}
	log.Println("backendBytes:::", backendBytes)
	log.Println("backendBytes2:::", string(backendBytes))

	noOfAllRequests++

	log.Println("allRequests:", allRequests)
	log.Println("lengthOfEachRequest:", lengthOfEachRequest)

	// mat.NewDense(noOfAllRequests, lengthOfEachRequest, allRequests)

}

func main() {
	var p string
	var r string
	flag.StringVar(
		&p,
		"p",
		"localhost:7777",
		"the IP:PORT pair to bind the proxy to.")
	flag.StringVar(
		&r,
		"r",
		"localhost:3333",
		"the IP:PORT pair to proxy connections to ie the remote server.")
	flag.Parse()

	{
		go proxyd.Run(r)
	}

	listener, err := net.Listen("tcp", p)
	if err != nil {
		err = errors.Wrapf(err, "Reverse failed to setup listener %v", p)
		log.Fatalf("\n%+v", err)
	}
	log.Println("Reverse Listening on " + p)

	for {
		reverseProxyConn, err := listener.Accept()
		if err != nil {
			err = errors.Wrapf(err, "Reverse failed to accept listener %v", p)
			log.Fatalf("\n%+v", err)
		}
		log.Printf("Accepted reverseProxyConnection %v\n", reverseProxyConn)
		go forward(reverseProxyConn, r)
	}
}

// z1 := []byte("komu")
// z2 := []byte("nomu")
// z3 := []byte("iomu")
// z4 := []byte("komr")
// z5 := []byte("komt")
// z6 := []byte("komx")
// z7 := []byte("komg")

// // X = np.array([z1, z2, z3, z4, z5, z6, z7])

// c := make([]float64, 0)
// for _, v := range z1 {
// 	c = append(c, float64(v))
// }
// for _, v := range z2 {
// 	c = append(c, float64(v))
// }
// for _, v := range z3 {
// 	c = append(c, float64(v))
// }
// for _, v := range z4 {
// 	c = append(c, float64(v))
// }
// for _, v := range z5 {
// 	c = append(c, float64(v))
// }
// for _, v := range z6 {
// 	c = append(c, float64(v))
// }
// for _, v := range z7 {
// 	c = append(c, float64(v))
// }

// X := mat.NewDense(7, 4, c)
