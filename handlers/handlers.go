package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/poty-tom/go-handson/models"
)

func HelthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello world")
}

func ArticleListHandler(w http.ResponseWriter, req *http.Request) {
	queryMap := req.URL.Query()
	var page int
	if p, ok := queryMap["page"]; ok && len(p) > 0 {
		var err error
		page, err = strconv.Atoi(p[0])
		if err != nil {
			http.Error(w, "invalid parameter", http.StatusBadRequest)
			return
		}
	} else {
		page = 1
	}

	resStr := fmt.Sprintf("Article List (page %d)\n", page)
	io.WriteString(w, resStr)
}

func PostArticleHandler(w http.ResponseWriter, req *http.Request) {
	var reqArticle models.Article
	if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
		http.Error(w, "fail to decode json\n", http.StatusBadRequest)
	}

	article := reqArticle

	json.NewEncoder(w).Encode(article)

}
