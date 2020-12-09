package client

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"tail-based-sampling/src/common"
)

func process(resBody *os.File) error {
    // func process(resBody io.Reader) error {
        linesPool := sync.Pool{New: func() interface{} {
            lines := make([]byte, 250*1024)
            return lines
        }}
        stringPool := sync.Pool{New: func() interface{} {
            lines := ""
            return lines
        }}
    
        reader := bufio.NewReader(resBody)
        batchNo := 0
        recordCounter := 0 
        for {
            buf := linesPool.Get().([]byte)
            n, err := reader.Read(buf)
            buf = buf[:n]
            if n == 0 {
                if err == nil { continue }
                if err == io.EOF { break }
                return err
            }
            nextUntillNewline, err := reader.ReadBytes('\n')
            if err != io.EOF {  buf = append(buf, nextUntillNewline...) }
    
            go func() {
                processChunk(buf, &linesPool, &stringPool, &batchNo, &recordCounter)
            }()
        }
    
        wg.Wait()
        return nil
}
    
func processChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, batchNo *int, recordCounter *int) {
    var wg2 sync.WaitGroup
    logs := stringPool.Get().(string)
    logs = string(chunk)
    
    linesPool.Put(chunk)
    logsSlice := strings.Split(logs, "\n")
    stringPool.Put(logs)
    chunkSize := 300
    n := len(logsSlice)
    noOfThread := n / chunkSize
    if n%chunkSize != 0 {
        noOfThread++
    }
    for i := 0; i < 2; i++ {
        wg2.Add(1)
        go func(s int, e int) {
            defer wg2.Done()
            for i := s; i < e; i++ {
                text := logsSlice[i]
                if len(text) == 0 {
                    continue
                }

                // fmt.Println(string(text))
                common.NewLineChan <- common.NewLine{Line: string(text), BatchNo: *batchNo}
            }
        }(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(logsSlice)))))
    }

    wg2.Wait()
    logsSlice = nil
}
