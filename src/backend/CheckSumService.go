package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"tail-based-sampling/src/common"
)

func SendCheckSum() {
	m := map[string]string{"e6ff221e869375a": "1757AA5015E69E84F47B08E55DB398A2"}
	for key, value := range m {
		m[key] = common.MD5(value)
	}
	mjson, _ := json.Marshal(m)
	mString := string(mjson)
	uploadPort := os.Getenv("UPLOAD_SERVER_PORT")
	res, err := http.PostForm("http://localhost:"+uploadPort+"/api/finished", url.Values{"result": {mString}})
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
