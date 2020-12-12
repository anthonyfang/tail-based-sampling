package backend

import (
	"os"
	"tail-based-sampling/src/common"
	"time"

	"github.com/gin-gonic/gin"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetParameterGetHandler(ctx *gin.Context) {

	port := ctx.Query("port")

	os.Setenv("UPLOAD_SERVER_PORT", port)

	ctx.String(200, "OK!")

	go func(port string) {
		time.Sleep(time.Millisecond * 10)
		common.ReadyChan <- port
	}(port)
}

func SetParameterPostHandler(ctx *gin.Context) {

	ctx.String(200, "Please use GET request instead!")
}
