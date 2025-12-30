// Package repository provides access to DB CRUD operations +analytics data
package repository

import (
	"context"
	"errors"

	"commentTree/internal/model"
)

type CommentRepository interface {
	Create(ctx context.Context, n *model.Comment) error
	GetAllRoot(ctx context.Context, limit, offset int, sort, order string) ([]model.Comment, error)
	DeleteByID(ctx context.Context, id int) error
	GetCommentByID(ctx context.Context, id int) (*model.Comment, error)
	GetChildrenByID(ctx context.Context, id int) ([]model.Comment, error)
	MarkAsDeletedByID(ctx context.Context, id int) error
	RunSearchQuery(ctx context.Context, query string) ([]model.Comment, error)
}

var ErrCommentNotFound error = errors.New("specified comment doesn't exist")
