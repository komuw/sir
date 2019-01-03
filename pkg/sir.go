package sir

import (
	"bytes"
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

func (backend backendType) String() string {
	names := []string{
		"Candidate",
		"Primary",
		"Secondary"}
	return names[backend]
}

type RequestsResponses struct {
	L                      sync.RWMutex
	NoOfAllRequests        int
	AllRequests            []float64
	LengthOfLargestRequest int
	RequestsSlice          [][]byte

	NoOfAllResponses        int
	AllResponses            []float64
	LengthOfLargestResponse int
	ResponsesSlice          [][]byte

	Backend backendType
}

func (reqResp *RequestsResponses) HandleRequest(requestBuf []byte) {
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	if reqResp.LengthOfLargestRequest < len(requestBuf) {
		reqResp.LengthOfLargestRequest = len(requestBuf)
	}
	reqResp.RequestsSlice = append(reqResp.RequestsSlice, requestBuf)
}

func (reqResp *RequestsResponses) HandleResponse(responseBuf []byte) {
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	if reqResp.LengthOfLargestResponse < len(responseBuf) {
		reqResp.LengthOfLargestResponse = len(responseBuf)
	}
	reqResp.ResponsesSlice = append(reqResp.ResponsesSlice, responseBuf)
}

func (reqResp *RequestsResponses) ClusterAndPlotRequests() {
	appendName := "Requests"
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	for k, v := range reqResp.RequestsSlice {
		diff := reqResp.LengthOfLargestRequest - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(NulByte), diff)
			v = append(v, pad...)
			reqResp.RequestsSlice[k] = v
		}
	}
	for _, eachRequest := range reqResp.RequestsSlice {
		for _, v := range eachRequest {
			reqResp.AllRequests = append(reqResp.AllRequests, float64(v))
		}
	}
	nclusters, X, err := GetClusters(reqResp.NoOfAllRequests, reqResp.LengthOfLargestRequest, reqResp.AllRequests, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Requests estimated number of clusters: %d\n", nclusters)

	proj := FindPCA(X, reqResp.LengthOfLargestRequest)
	err = PlotResultsPCA(reqResp.NoOfAllRequests, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}

func (reqResp *RequestsResponses) ClusterAndPlotResponses() {
	appendName := "Responses"
	reqResp.L.Lock()
	defer reqResp.L.Unlock()

	for k, v := range reqResp.ResponsesSlice {
		diff := reqResp.LengthOfLargestResponse - len(v)
		if diff != 0 {
			pad := bytes.Repeat([]byte(NulByte), diff)
			v = append(v, pad...)
			reqResp.ResponsesSlice[k] = v
		}
	}
	for _, eachResponse := range reqResp.ResponsesSlice {
		for _, v := range eachResponse {
			reqResp.AllResponses = append(reqResp.AllResponses, float64(v))
		}
	}
	nclusters, X, err := GetClusters(reqResp.NoOfAllResponses, reqResp.LengthOfLargestResponse, reqResp.AllResponses, 3.0, 1.0, false, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
	log.Printf("Responses stimated number of clusters: %d\n", nclusters)

	proj := FindPCA(X, reqResp.LengthOfLargestResponse)
	err = PlotResultsPCA(reqResp.NoOfAllResponses, proj, nclusters, appendName)
	if err != nil {
		log.Fatalf("\n%+v", err)
	}
}
