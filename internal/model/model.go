// Package model provides data structures for repository layer DB-interaction
package model

import "time"

const (
	ByContent = "text_content"
	ByAuthor  = "author"
	ByCreated = "created"
	OrderASC  = "ascending"
	OrderDESC = "descending"
)

type DBComment struct {
	ID        int
	ParentID  *int
	Text      string
	CreatedAt time.Time
	DeletedAt *time.Time
	Author    string
}

type RootRequest struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Sort  string `form:"sort"`
	Order string `form:"order"`
}
