package repositories_test

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/poty-tom/go-handson/models"
	"github.com/poty-tom/go-handson/repositories"
)

func TestSelectArticleDetail(t *testing.T) {
	dbUser := "mysql"
	dbPassword := "password"
	dbDatabase := "sample"
	dbConn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPassword, dbDatabase)

	db, err := sql.Open("mysql", dbConn)

	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	expected := models.Article{
		ID:       1,
		Title:    "firstPost",
		Contents: "This is my first blog",
		UserName: "saki",
		NiceNum:  2,
	}

	got, err := repositories.SelectArticleDetail(db, expected.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != expected.ID {
		t.Errorf("ID is incorrect")
	}

	if got.Title != expected.Title {
		t.Errorf("Title is incorrect")
	}

}
