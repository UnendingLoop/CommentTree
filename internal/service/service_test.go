package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"commentTree/internal/model"
	"commentTree/internal/repository"
)

type mockRepo struct {
	getByIDFn         func(ctx context.Context, id int) (*model.DBComment, error)
	createFn          func(ctx context.Context, c *model.CommentCreateData) (*model.DBComment, error)
	getAllRootFn      func(ctx context.Context, limit, offset int, sort, order string) ([]model.DBComment, error)
	getWithChildrenFn func(ctx context.Context, id int) ([]model.DBComment, error)
	markDeletedFn     func(ctx context.Context, id int) error
	deleteFn          func(ctx context.Context, id int) error
	runSearchFn       func(ctx context.Context, query string) ([]model.DBComment, error)
}

func (m *mockRepo) GetCommentByID(ctx context.Context, id int) (*model.DBComment, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) Create(ctx context.Context, c *model.CommentCreateData) (*model.DBComment, error) {
	return m.createFn(ctx, c)
}

func (m *mockRepo) GetAllRoot(ctx context.Context, limit, offset int, sort, order string) ([]model.DBComment, error) {
	return m.getAllRootFn(ctx, limit, offset, sort, order)
}

func (m *mockRepo) GetCommentWithChildrenByID(ctx context.Context, id int) ([]model.DBComment, error) {
	return m.getWithChildrenFn(ctx, id)
}

func (m *mockRepo) MarkAsDeletedByID(ctx context.Context, id int) error {
	return m.markDeletedFn(ctx, id)
}

func (m *mockRepo) DeleteByID(ctx context.Context, id int) error {
	return m.deleteFn(ctx, id)
}

func (m *mockRepo) RunSearchQuery(ctx context.Context, query string) ([]model.DBComment, error) {
	return m.runSearchFn(ctx, query)
}

/*
	CREATE COMMENT
*/

func TestCreateComment_OK(t *testing.T) {
	repo := &mockRepo{
		createFn: func(ctx context.Context, c *model.CommentCreateData) (*model.DBComment, error) {
			return &model.DBComment{ID: 1, Text: c.Text}, nil
		},
	}

	svc := NewCommentService(repo)

	res, err := svc.CreateComment(context.Background(), &model.CommentCreateData{
		Text: "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ID != 1 {
		t.Fatalf("unexpected id")
	}
}

func TestCreateComment_ParentNotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return nil, repository.ErrCommentNotFound
		},
	}

	svc := NewCommentService(repo)
	parentID := 10

	_, err := svc.CreateComment(context.Background(), &model.CommentCreateData{
		ParentID: &parentID,
	})

	if !errors.Is(err, ErrParentNotFound) {
		t.Fatalf("expected ErrParentNotFound")
	}
}

func TestCreateComment_ParentDeleted(t *testing.T) {
	now := time.Now()

	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return &model.DBComment{ID: id, DeletedAt: &now}, nil
		},
	}

	svc := NewCommentService(repo)
	parentID := 5

	_, err := svc.CreateComment(context.Background(), &model.CommentCreateData{
		ParentID: &parentID,
	})

	if !errors.Is(err, ErrParentDeleted) {
		t.Fatalf("expected ErrParentDeleted")
	}
}

/*
	GET ALL ROOT COMMENTS
*/

func TestGetAllRootComments_OK(t *testing.T) {
	repo := &mockRepo{
		getAllRootFn: func(ctx context.Context, limit, offset int, sort, order string) ([]model.DBComment, error) {
			return []model.DBComment{
				{ID: 1, Text: "root"},
			}, nil
		},
	}

	svc := NewCommentService(repo)

	res, err := svc.GetAllRootComments(context.Background(), &model.RootRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 comment")
	}
}

/*
	GET COMMENT WITH CHILDREN
*/

func TestGetCommentWithChildren_InvalidID(t *testing.T) {
	svc := NewCommentService(&mockRepo{})

	_, err := svc.GetCommentWithChildren(context.Background(), 0)
	if !errors.Is(err, ErrIncorrectID) {
		t.Fatalf("expected ErrIncorrectID")
	}
}

func TestGetCommentWithChildren_ParentNotFound(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return nil, repository.ErrCommentNotFound
		},
	}

	svc := NewCommentService(repo)

	_, err := svc.GetCommentWithChildren(context.Background(), 1)
	if !errors.Is(err, ErrParentNotFound) {
		t.Fatalf("expected ErrParentNotFound")
	}
}

func TestGetCommentWithChildren_OK(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return &model.DBComment{ID: id}, nil
		},
		getWithChildrenFn: func(ctx context.Context, id int) ([]model.DBComment, error) {
			return []model.DBComment{
				{ID: id},
				{ID: 2, ParentID: &id},
			}, nil
		},
	}

	svc := NewCommentService(repo)

	res, err := svc.GetCommentWithChildren(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(res) != 1 {
		t.Fatalf("expected root node")
	}
}

/*
	DELETE COMMENT
*/

func TestDeleteComment_SoftDelete(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return &model.DBComment{ID: id}, nil // <- комментарий существует
		},
		markDeletedFn: func(ctx context.Context, id int) error {
			return nil
		},
	}

	svc := NewCommentService(repo)

	if err := svc.DeleteCommentByID(context.Background(), 1, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteComment_HardDelete(t *testing.T) {
	repo := &mockRepo{
		getByIDFn: func(ctx context.Context, id int) (*model.DBComment, error) {
			return &model.DBComment{ID: id}, nil
		},
		deleteFn: func(ctx context.Context, id int) error {
			return nil
		},
	}

	svc := NewCommentService(repo)

	if err := svc.DeleteCommentByID(context.Background(), 1, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

/*
	SEARCH
*/

func TestRunCommentSearchQuery_Empty(t *testing.T) {
	svc := NewCommentService(&mockRepo{})

	res, err := svc.RunCommentSearchQuery(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if res != nil {
		t.Fatalf("expected nil result")
	}
}

func TestRunCommentSearchQuery_OK(t *testing.T) {
	repo := &mockRepo{
		runSearchFn: func(ctx context.Context, query string) ([]model.DBComment, error) {
			return []model.DBComment{
				{ID: 1, Text: "match"},
			}, nil
		},
	}

	svc := NewCommentService(repo)

	res, err := svc.RunCommentSearchQuery(context.Background(), "match")
	if err != nil {
		t.Fatalf("unexpected error")
	}
	if len(res) != 1 {
		t.Fatalf("expected 1 result")
	}
}
