// Package model provides data structures for the whole application commentTree
package model

import "time"

const (
	ByContent = "text_content"
	ByAuthor  = "author"
	ByCreated = "created"
	OrderASC  = "ascending"
	OrderDESC = "descending"
)

type Comment struct {
	ID        int        `json:"id,omitempty"`
	ParentID  *int       `json:"parent_id,omitempty"`
	Text      string     `json:"content"`
	CreatedAt time.Time  `json:"created_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	Author    string     `json:"author,omitempty"`
}

type RootRequest struct {
	Page  int    `form:"page"`
	Limit int    `form:"limit"`
	Sort  string `form:"sort"`
	Order string `form:"order"`
}
