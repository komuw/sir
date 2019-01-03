package sir

import "sync"

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
