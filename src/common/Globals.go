package common

const (
	IS_DEBUG       = true
	DEBUG_TRACE_ID = "x"
	IS_PRINTOUT    = true
)

//var clientHosts = []string{"http://localhost:8000"}
var ClientHosts = []string{"http://localhost:8000", "http://localhost:8001"}
var BackendHosts = []string{"http://localhost:8002"}

// WS Interfaces
type Payload struct {
	Action  string      `json:"action"`
	ID      string      `json:"id"`
	Records interface{} `json:"records"`
}

func (p *Payload) GetWrongTraceGen(traceid string) {
	p.Action = "GetWrongTrace"
	p.ID = traceid
	p.Records = nil
}

func (p *Payload) ReturnWrongTraceGen(traceid string, data []string) {
	p.Action = "ReturnWrongTrace"
	p.ID = traceid
	p.Records = data
}

func (p *Payload) SetWrongTraceIDGen(traceid string, data []string) {
	p.Action = "SetWrongTraceID"
	p.ID = traceid
	p.Records = data
}

func (p *Payload) SendFinishGen(port string) {
	p.Action = "SendFinished"
	p.ID = port
	p.Records = []string{}
}

type NewLine struct {
	Line    string
	BatchNo int
}

// Channels
// ------------ Clients ----------------- //

var CacheChan = make(chan string, 1)

var NewLineChan = make(chan NewLine, 2000000)

// PostTraceChan is a channel for sending/receiving the signal
var PostTraceChan = make(chan string)

// ------------ Backend ----------------- //

var BackendChan = make(chan string)

var BatchReceivedCountChan = make(chan int, 10)

// ------------ Common ----------------- //
var FinishedChan = make(chan string)

var ClientRecvWSChan = make(chan string)

var ClientSendWSChan = make(chan *RecordTemplate)

var ServerRecvWSChan = make(chan *RecordTemplate)

var ServerSendWSChan = make(chan interface{})

var ReceivedTraceInfoChan = make(chan string)
