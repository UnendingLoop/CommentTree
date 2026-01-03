package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/UnendingLoop/CommentTree/internal/model"
	"github.com/UnendingLoop/CommentTree/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

/*
	MOCK SERVICE
*/

type mockService struct {
	createFn     func(ctx context.Context, c *model.CommentCreateData) (*service.APPComment, error)
	getAllRootFn func(ctx context.Context, req *model.RootRequest) ([]service.APPComment, error)
	getByIDFn    func(ctx context.Context, id int) ([]service.APPComment, error)
	deleteFn     func(ctx context.Context, id int, soft bool) error
	searchFn     func(ctx context.Context, q string) ([]service.APPComment, error)
}

func (m *mockService) CreateComment(ctx context.Context, c *model.CommentCreateData) (*service.APPComment, error) {
	return m.createFn(ctx, c)
}

func (m *mockService) GetAllRootComments(ctx context.Context, req *model.RootRequest) ([]service.APPComment, error) {
	return m.getAllRootFn(ctx, req)
}

func (m *mockService) GetCommentWithChildren(ctx context.Context, id int) ([]service.APPComment, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockService) DeleteCommentByID(ctx context.Context, id int, soft bool) error {
	return m.deleteFn(ctx, id, soft)
}

func (m *mockService) RunCommentSearchQuery(ctx context.Context, q string) ([]service.APPComment, error) {
	return m.searchFn(ctx, q)
}

/*
	HELPERS
*/

func setupRouter(handler *CommentsHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/ping", ginext.HandlerFunc(handler.SimplePinger))
	r.POST("/comments", ginext.HandlerFunc(handler.Create))
	r.GET("/comments", ginext.HandlerFunc(handler.GetAllRootComments))
	r.GET("/comments/:id", ginext.HandlerFunc(handler.GetCommentWithChildren))
	r.DELETE("/comments/:id", ginext.HandlerFunc(handler.DeleteComment))
	r.GET("/search", ginext.HandlerFunc(handler.RunSearch))

	return r
}

/*
	SIMPLE PING
*/

func TestSimplePinger(t *testing.T) {
	h := NewCommentHandlers(&mockService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

/*
	CREATE
*/

func TestCreate_OK(t *testing.T) {
	svc := &mockService{
		createFn: func(ctx context.Context, c *model.CommentCreateData) (*service.APPComment, error) {
			return &service.APPComment{ID: 1}, nil
		},
	}

	h := NewCommentHandlers(svc)
	r := setupRouter(h)

	body, _ := json.Marshal(map[string]string{"content": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestCreate_BindError(t *testing.T) {
	h := NewCommentHandlers(&mockService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewReader([]byte("bad json")))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

/*
	GET ROOT COMMENTS
*/

func TestGetAllRootComments_OK(t *testing.T) {
	svc := &mockService{
		getAllRootFn: func(ctx context.Context, req *model.RootRequest) ([]service.APPComment, error) {
			return []service.APPComment{{ID: 1, Text: "bla"}, {ID: 2, Text: "bla"}, {ID: 3, Text: "bla"}}, nil
		},
	}

	h := NewCommentHandlers(svc)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/comments?page=1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

/*
	GET COMMENT WITH CHILDREN
*/

func TestGetCommentWithChildren_OK(t *testing.T) {
	svc := &mockService{
		getByIDFn: func(ctx context.Context, id int) ([]service.APPComment, error) {
			return []service.APPComment{{ID: id}}, nil
		},
	}

	h := NewCommentHandlers(svc)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/comments/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestGetCommentWithChildren_InvalidID(t *testing.T) {
	h := NewCommentHandlers(&mockService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/comments/abc", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

/*
	DELETE
*/

func TestDeleteComment_Soft_OK(t *testing.T) {
	svc := &mockService{
		deleteFn: func(ctx context.Context, id int, soft bool) error {
			if !soft {
				t.Fatalf("expected soft delete")
			}
			return nil
		},
	}

	h := NewCommentHandlers(svc)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/comments/1?mode=soft", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204")
	}
}

func TestDeleteComment_InvalidMode(t *testing.T) {
	h := NewCommentHandlers(&mockService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/comments/1?mode=xxx", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

/*
	SEARCH
*/

func TestRunSearch_OK(t *testing.T) {
	svc := &mockService{
		searchFn: func(ctx context.Context, q string) ([]service.APPComment, error) {
			return []service.APPComment{{ID: 1}}, nil
		},
	}

	h := NewCommentHandlers(svc)
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200")
	}
}

func TestRunSearch_EmptyQuery(t *testing.T) {
	h := NewCommentHandlers(&mockService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/search?q=", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400")
	}
}

/*
	ERROR MAPPING
*/

func TestErrorCodeDefiner(t *testing.T) {
	tests := []struct {
		err  error
		code int
	}{
		{service.ErrParentNotFound, 404},
		{service.ErrIncorrectID, 400},
		{service.ErrParentDeleted, 409},
		{service.ErrCommon500, 500},
		{errors.New("unknown"), 500},
	}

	for _, tt := range tests {
		if got := errorCodeDefiner(tt.err); got != tt.code {
			t.Fatalf("expected %d, got %d", tt.code, got)
		}
	}
}
