package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/poty-tom/go-handson/models"
)

// [GET: /]
// ヘルスチェック用
func HelthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello world")
}

// [GET: /article/list/page={number}]
func ArticleListHandler(w http.ResponseWriter, req *http.Request) {
	queryMap := req.URL.Query() //この時点でクエリパラメータをマップに格納
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

	log.Println(page)
	articleList := []models.Article{models.Article1}
	json.NewEncoder(w).Encode(articleList)
}

// [GET: /article/{id}]
// {id}値に応じたArticleデータ取得
func ArticleDetailHandler(w http.ResponseWriter, req *http.Request) {
	articleID, err := strconv.Atoi(mux.Vars(req)["id"]) // パスパラメータを取得
	if err != nil {
		http.Error(w, "invalid path parameter", http.StatusBadRequest)
		return
	}

	// TODO: ArticleIDを検索
	log.Println(articleID)
	article := models.Article1
	json.NewEncoder(w).Encode(article)
}

// [POST: /article]
// Articleデータ作成
func PostArticleHandler(w http.ResponseWriter, req *http.Request) {
	var reqArticle models.Article
	if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
		http.Error(w, "fail to decode json\n", http.StatusBadRequest)
	}

	article := reqArticle

	json.NewEncoder(w).Encode(article)

}

// [POST: /article/nice]
func PostNiceHandler(w http.ResponseWriter, req *http.Request) {
	var reqArticle models.Article
	if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
		http.Error(w, "fail to decode json\n", http.StatusBadRequest)
	}
	// TODO: ArticleにNiceする
	article := reqArticle
	json.NewEncoder(w).Encode(&article)
}

// [POST: /comment]
// commentデータ作成
func PostCommentHandler(w http.ResponseWriter, req *http.Request) {
	var reqComment models.Comment
	if err := json.NewDecoder(req.Body).Decode(&reqComment); err != nil {
		http.Error(w, "fail to decode json\n", http.StatusBadRequest)
	}

	comment := reqComment
	json.NewEncoder(w).Encode(comment)
}