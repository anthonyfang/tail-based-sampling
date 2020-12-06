package common

// ------------ Clients ----------------- //

var CacheChan = make(chan string, 1)

var NewLineChan = make(chan string, 200000)

// PostTraceChan is a channel for sending/receiving the signal
var PostTraceChan = make(chan string)

// ------------ Backend ----------------- //

var BackendChan = make(chan string)

var BatchReceivedCountChan = make(chan int, 10)

var GenCheckSumToQueueChan = make(chan string)

// ------------ Common ----------------- //
var FinishedChan = make(chan string)
