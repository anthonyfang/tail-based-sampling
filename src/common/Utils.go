package common

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"strings"
)

// GetEnvDefault is using for getting enviroment variable with default value
func GetEnvDefault(key string, defVal string) string {
	val, ex := os.LookupEnv(key)
	if !ex {
		return defVal
	}
	return val
}

// MD5 is using for generating checkSum
func MD5(spans string) string {
	h := md5.New()
	io.WriteString(h, spans)
	sum := h.Sum(nil)
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}
