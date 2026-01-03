// Package repository provides access to DB CRUD operations
package repository

import (
	"context"
	"errors"

	"github.com/UnendingLoop/CommentTree/internal/model"
)

type CommentRepository interface {
	Create(ctx context.Context, n *model.CommentCreateData) (*model.DBComment, error)
	GetAllRoot(ctx context.Context, limit, offset int, sort, order string) ([]model.DBComment, error)
	DeleteByID(ctx context.Context, id int) error
	GetCommentByID(ctx context.Context, id int) (*model.DBComment, error)
	GetCommentWithChildrenByID(ctx context.Context, id int) ([]model.DBComment, error)
	MarkAsDeletedByID(ctx context.Context, id int) error
	RunSearchQuery(ctx context.Context, query string) ([]model.DBComment, error)
}

var ErrCommentNotFound error = errors.New("specified comment doesn't exist")
