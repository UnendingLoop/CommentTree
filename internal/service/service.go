// Package service provides business-logic for the app
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"commentTree/internal/model"
	"commentTree/internal/mwlogger"
	"commentTree/internal/repository"
)

var (
	ErrCommon500      error = errors.New("something went wrong. Try again later") // 500
	ErrIncorrectQuery error = errors.New("incorrect query parameters")            // 400
	ErrParentNotFound error = errors.New("specified parent comment ID not found") // 404
	ErrParentDeleted  error = errors.New("specified parent ID is deleted")        // 422
	ErrIncorrectID    error = errors.New("incorrect comment ID")                  // 422
)

type CommentService interface {
	CreateComment(ctx context.Context, comment *model.CommentCreateData) (*APPComment, error)
	GetAllRootComments(ctx context.Context, req *model.RootRequest) ([]APPComment, error)
	GetCommentWithChildren(ctx context.Context, id int) ([]APPComment, error)
	DeleteCommentByID(ctx context.Context, id int, isSoftDelete bool) error
	RunCommentSearchQuery(ctx context.Context, query string) ([]APPComment, error)
}

type CService struct {
	repo repository.CommentRepository
}

func NewCommentService(commentRep repository.CommentRepository) CommentService {
	return &CService{repo: commentRep}
}

func (c CService) CreateComment(ctx context.Context, comment *model.CommentCreateData) (*APPComment, error) {
	logger := mwlogger.LoggerFromContext(ctx)
	// если указан родитель, проверяем его в базе
	if comment.ParentID != nil {
		parent, err := c.repo.GetCommentByID(ctx, *comment.ParentID)
		if err != nil {
			switch {
			case errors.Is(err, repository.ErrCommentNotFound):
				return nil, ErrParentNotFound
			default:
				logger.Error().Err(err).Msg("Failed to check parent before creating new comment in DB")
				return nil, ErrCommon500
			}
		}

		if parent.DeletedAt != nil { // оставлять коммент мягко удаленному родителю запрещено
			return nil, ErrParentDeleted
		}
	}

	res, err := c.repo.Create(ctx, comment)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to create new comment")
		return nil, ErrCommon500
	}

	return convertToAPPComment(res), nil
}

func (c CService) GetAllRootComments(ctx context.Context, req *model.RootRequest) ([]APPComment, error) {
	logger := mwlogger.LoggerFromContext(ctx)
	validateRequest(req)
	offset := (req.Page - 1) * req.Limit

	res, err := c.repo.GetAllRoot(ctx, req.Limit, offset, req.Sort, req.Order)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to fetch all root comments from DB")
		return nil, ErrCommon500
	}

	return compileToAPPCommentTree(res, nil), nil
}

func (c CService) GetCommentWithChildren(ctx context.Context, id int) ([]APPComment, error) {
	logger := mwlogger.LoggerFromContext(ctx)
	if id <= 0 {
		return nil, ErrIncorrectID
	}
	// проверяем существует ли такой родитель
	_, err := c.repo.GetCommentByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrCommentNotFound):
			return nil, ErrParentNotFound
		default:
			logger.Error().Err(err).Msg("Failed to check parent before fetching its children")
			return nil, ErrCommon500
		}
	}

	res, err := c.repo.GetCommentWithChildrenByID(ctx, id)
	if err != nil {
		logger.Error().Err(err).Msg(fmt.Sprintf("Failed to fetch children for comment %q from DB", id))
		return nil, ErrCommon500
	}

	return compileToAPPCommentTree(res, &id), nil
}

func (c CService) DeleteCommentByID(ctx context.Context, id int, isSoftDelete bool) error {
	logger := mwlogger.LoggerFromContext(ctx)
	if id <= 0 {
		return ErrIncorrectID
	}

	// проверяем существует ли такой коммент
	_, err := c.repo.GetCommentByID(ctx, id)
	if err != nil {
		switch {
		case !errors.Is(err, repository.ErrCommentNotFound):
			logger.Error().Err(err).Msg("Failed to check comment existence before deleting/hiding")
			return ErrCommon500
		default:
			return err
		}
	}

	// определяем режим удаления
	switch isSoftDelete {
	case true:
		if err := c.repo.MarkAsDeletedByID(ctx, id); err != nil {
			logger.Error().Err(err).Msg("Failed to mark comment as deleted")
			return ErrCommon500
		}
	default:
		if err := c.repo.DeleteByID(ctx, id); err != nil {
			logger.Error().Err(err).Msg("Failed to delete comment")
			return ErrCommon500
		}
	}
	return nil
}

func (c CService) RunCommentSearchQuery(ctx context.Context, query string) ([]APPComment, error) {
	if query == "" {
		return nil, nil
	}
	logger := mwlogger.LoggerFromContext(ctx)

	res, err := c.repo.RunSearchQuery(ctx, query)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to run search query in DB")
		return nil, ErrCommon500
	}
	return convertSearchResults(res), nil
}

func validateRequest(req *model.RootRequest) {
	// Обрабатываем пустые значения, присваиваем дефолты если надо
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 30
	}
	if req.Sort == "" {
		req.Sort = model.ByCreated
	}
	if req.Order == "" {
		req.Order = model.OrderDESC
	}

	// Валидируем непустое поле типа сортировки
	req.Sort = strings.ToLower(req.Sort)
	req.Sort = strings.TrimSpace(req.Sort)
	switch {
	case strings.Contains(model.ByAuthor, req.Sort):
		req.Sort = "author"
	case strings.Contains(model.ByContent, req.Sort):
		req.Sort = "content"
	case strings.Contains(model.ByCreated, req.Sort):
		req.Sort = "created_at"
	default:
		req.Sort = "created_at" // по дефолту ставим сортировку по времени создания
	}

	// Валадируем непустой порядок
	req.Order = strings.ToLower(req.Order)
	req.Order = strings.TrimSpace(req.Order)
	switch {
	case strings.Contains(model.OrderASC, req.Order):
		req.Order = "ASC"
	case strings.Contains(model.OrderDESC, req.Order):
		req.Order = "DESC"
	default:
		req.Order = "ASC"
	}
}
