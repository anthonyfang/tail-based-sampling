package client

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"tail-based-sampling/src/common"
	"tail-based-sampling/src/trace"
	"time"

	"github.com/gin-gonic/gin"
)

var isRunning = false
var counter uint64 = 0

func getURL(port string) string {
	var url string
	port = "8080"
	switch currentServerPort := common.GetEnvDefault("SERVER_PORT", ""); currentServerPort {
	case "8000":
		url = fmt.Sprintf("http://localhost:%v/trace1-4G.data", port)
	case "8001":
		url = fmt.Sprintf("http://localhost:%v/trace2-4G.data", port)
	default:
		url = ""
	}

	return url
}

// SetParameterPostHandler is use for handling the SetParameterHandler endpoint
func SetParameterPostHandler(ctx *gin.Context) {
	type Request struct {
		Port string `json:"port"`
	}
	var body Request

	err := ctx.BindJSON(&body)
	if err != nil {
		fmt.Println(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot parse Body JSON"})
	}
	fmt.Printf("info: %#v\n", body)

	ctx.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
	})

	os.Setenv("UPLOAD_SERVER_PORT", body.Port)

	ctx.String(200, "Please use Get Request! Upload server port is: %v", body.Port)
}

// SetParameterGetHandler is use for handling the SetParameterHandler endpoint
func SetParameterGetHandler(ctx *gin.Context) {
	port := ctx.Query("port")

	os.Setenv("UPLOAD_SERVER_PORT", port)
	serverPort := common.GetEnvDefault("SERVER_PORT", "8002")

	if serverPort != "8002" && !isRunning {
		isRunning = true

		go func(port string) {
			time.Sleep(time.Millisecond * 10)
			common.ReadyChan <- port
		}(serverPort)
	}

	ctx.String(200, "Upload server port is: %v", port)
}

// fetchData is use for fetching the data file from datasource server
func fetchData(url string) {
	startTime := time.Now()
	fmt.Println("################# : fetchingData", startTime)
	if url != "" {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalln(err)
		}

		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, bufio.MaxScanTokenSize)

		for scanner.Scan() {
			recordString := scanner.Text()
			common.NewLineChan <- common.NewLine{Line: recordString, BatchNo: int(batchNo)}
		}

		close(common.NewLineChan)
		for msg := range common.FinishedChan {
			if msg == "readline" {
				TimeChan <- timeWindowEnd + 1
				common.Wg.Wait()
				close(TimeChan)
			}

			if msg == "timeWindow" {
				fmt.Println("xxxxxxxxxxxxxxxxxx: pushed ", counter)
				go postFinishSignal()

				fmt.Println("################# : fetchingData END", time.Now())
				fmt.Println("################# : fetchingData Total Elapsed Time: ", time.Since(startTime))
				// return
			}
		}
	}
}

func postFinishSignal() {
	// var payload = new(common.Payload)
	// payload.SendFinishGen(common.GetEnvDefault("SERVER_PORT", "8002"))

	// msg, _ := json.Marshal(payload)

	payload := &trace.PayloadMessage{
		Action:  "SendFinished",
		ID:      "0",
		Records: []string{},
	}
	if err := (*gRPCstream).Send(payload); err != nil {
		log.Fatal(err)
	}
}
