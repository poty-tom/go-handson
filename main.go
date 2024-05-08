package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/poty-tom/go-handson/handlers"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handlers.HelthCheck).Methods(http.MethodGet)
	r.HandleFunc("/article/list", handlers.ArticleListHandler).Methods(http.MethodGet)
	r.HandleFunc("/article", handlers.PostArticleHandler).Methods(http.MethodPost)
	r.HandleFunc("/artcile/{id:[0-9]+}", handlers.ArticleDetailHandler).Methods(http.MethodPost)
	r.HandleFunc("/article/nice", handlers.PostNiceHandler).Methods(http.MethodPost)
	r.HandleFunc("/coment", handlers.PostCommentHandler).Methods(http.MethodPost)
	log.Println("server start at port: 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
