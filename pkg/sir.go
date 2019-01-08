package sir

import (
	"bytes"
	"fmt"
	"log"
	"sync"
)

const NulByte = "\x00"

type backendType int

const (
	Candidate backendType = iota
	Primary
	Secondary
)

func (bt backendType) String() string {
	names := []string{
		"Candidate",
		"Primary",
		"Secondary"}
	return names[bt]
}

type Backend struct {
	Type backendType
	Addr string
}

func (b Backend) String() string {
	return fmt.Sprintf("%v(%v)", b.Type, b.Addr)
}

type RequestsResponse struct {
	L                      sync.RWMutex
	NoOfAllRequests        int
	AllRequests            []float64
	LengthOfLargestRequest int
	RequestsSlice          [][]byte

	NoOfAllResponses        int
	AllResponses            []float64
	LengthOfLargestResponse int
	ResponsesSlice          [][]byte

	Backend
}

func (reqResp *RequestsResponse) HandleRequest(requestBuf []byte) {
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	if reqResp.LengthOfLargestRequest < len(requestBuf) {
		reqResp.LengthOfLargestRequest = len(requestBuf)
	}
	reqResp.RequestsSlice = append(reqResp.RequestsSlice, requestBuf)
	reqResp.NoOfAllRequests++
}

func (reqResp *RequestsResponse) HandleResponse(responseBuf []byte) {
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	if reqResp.LengthOfLargestResponse < len(responseBuf) {
		reqResp.LengthOfLargestResponse = len(responseBuf)
	}
	reqResp.ResponsesSlice = append(reqResp.ResponsesSlice, responseBuf)
	reqResp.NoOfAllResponses++
}

// TODO: this should return error
func (reqResp *RequestsResponse) ClusterAndPlotRequests() {
	appendName := "Requests:" + fmt.Sprint(reqResp.Backend)

	for k, v := range reqResp.RequestsSlice {
		// eliminate race condition of runtime.slicecopy
		bufCopy := make([]byte, len(v))
		copy(bufCopy, v)

		diff := reqResp.LengthOfLargestRequest - len(bufCopy)
		if diff != 0 {
			pad := bytes.Repeat([]byte(NulByte), diff)
			bufCopy = append(bufCopy, pad...)
			reqResp.RequestsSlice[k] = bufCopy
		}
	}
	for _, eachRequest := range reqResp.RequestsSlice {
		for _, v := range eachRequest {
			reqResp.AllRequests = append(reqResp.AllRequests, float64(v))
		}
	}
	log.Printf("lengthOfLargestRequest for backend %v %v", reqResp.Backend, reqResp.LengthOfLargestRequest)
	log.Printf("noOfAllRequests for backend %v %v ", reqResp.Backend, reqResp.NoOfAllRequests)
	log.Printf("len(reqResp.AllRequests) for backend %v %v ", reqResp.Backend, len(reqResp.AllRequests))
	nclusters, X, err := GetClusters(reqResp.NoOfAllRequests, reqResp.LengthOfLargestRequest, reqResp.AllRequests, 3.0, 1.0, false)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Requests estimated number of clusters for backend %v: %d \n", reqResp.Backend, nclusters)

	proj := FindPCA(X, reqResp.LengthOfLargestRequest)
	err = PlotResultsPCA(reqResp.NoOfAllRequests, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

// TODO: this should return error
func (reqResp *RequestsResponse) ClusterAndPlotResponses() {
	appendName := "Responses:" + fmt.Sprint(reqResp.Backend)
	for k, v := range reqResp.ResponsesSlice {
		// eliminate race condition of runtime.slicecopy
		bufCopy := make([]byte, len(v))
		copy(bufCopy, v)

		diff := reqResp.LengthOfLargestResponse - len(bufCopy)
		if diff != 0 {
			pad := bytes.Repeat([]byte(NulByte), diff)
			bufCopy = append(bufCopy, pad...)
			reqResp.ResponsesSlice[k] = bufCopy
		}
	}
	for _, eachResponse := range reqResp.ResponsesSlice {
		for _, v := range eachResponse {
			reqResp.AllResponses = append(reqResp.AllResponses, float64(v))
		}
	}
	log.Printf("lengthOfLargestResponse for backend %v %v", reqResp.Backend, reqResp.LengthOfLargestResponse)
	nclusters, X, err := GetClusters(reqResp.NoOfAllResponses, reqResp.LengthOfLargestResponse, reqResp.AllResponses, 3.0, 1.0, false)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Responses estimated number of clusters for backend %v: %d\n", reqResp.Backend, nclusters)

	proj := FindPCA(X, reqResp.LengthOfLargestResponse)
	err = PlotResultsPCA(reqResp.NoOfAllResponses, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}
