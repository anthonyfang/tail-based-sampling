package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"tail-based-sampling/src/common"
)

func sendCheckSum(m map[string]string) {
	resultQueueLocker.Lock()
	mjson, _ := json.Marshal(m)
	mString := string(mjson)
	uploadPort := common.GetEnvDefault("UPLOAD_SERVER_PORT", "8080")
	res, err := http.PostForm("http://localhost:"+uploadPort+"/api/finished", url.Values{"result": {mString}})
	resultQueueLocker.Unlock()

	if err != nil {
		log.Fatal(err)
	}
	robots, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("response: %s", robots)
	fmt.Println("")
}
