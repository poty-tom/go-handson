package main

import (
	"log"
	"net/http"

	"github.com/poty-tom/go-handson/handlers"
)

func main() {

	http.HandleFunc("/", handlers.HelthCheck)

	log.Println("server start at port: 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
