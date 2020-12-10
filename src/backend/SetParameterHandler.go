package backend

import (
	"os"

	"github.com/gin-gonic/gin"
)

// SetWrongTraceIDHandler is use for handle the setWrongTraceId endpoint
func SetParameterGetHandler(ctx *gin.Context) {

	port := ctx.Query("port")

	os.Setenv("UPLOAD_SERVER_PORT", port)

	go processing()

	ctx.String(200, "OK!")
}

func SetParameterPostHandler(ctx *gin.Context) {

	ctx.String(200, "Please use GET request instead!")
}
