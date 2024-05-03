package handlers

import (
	"io"
	"net/http"
)

func HelthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello world")
}
