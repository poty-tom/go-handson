package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poty-tom/go-handson/handlers"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HelthCheck)
	r.HandleFunc("/article", handlers.ArticleListHandler).Queries("page", "{page}")

	log.Println("server start at port: 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
