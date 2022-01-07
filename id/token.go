package id

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"time"
)

func GenerateToken() string {
	currentTime := time.Now().Unix()
	h := md5.New()
	io.WriteString(h, strconv.FormatInt(currentTime, 10))
	token := fmt.Sprintf("%x", h.Sum(nil))
	return token
}