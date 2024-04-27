package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	handler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "hello world\n")
	}

	http.HandleFunc("/", handler)

	log.Println("server start at port: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
