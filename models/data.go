package models

import "time"

// Comment sample data
var (
	Comment1 = Comment{
		CommentID: 1,
		ArticleID: 1,
		Message:   "test comemnt1",
		CreatedAt: time.Now(),
	}

	Comment2 = Comment{
		CommentID: 2,
		ArticleID: 1,
		Message:   "second comment",
		CreatedAt: time.Now(),
	}
)

// Article sample data
var (
	Article1 = Article{
		ID:          1,
		Title:       "First Article",
		UserName:    "This is the test article",
		NiceNum:     1,
		CommentList: []Comment{Comment1, Comment2},
		CreatedAt:   time.Now(),
	}
)
