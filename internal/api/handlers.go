// Package api provides methods for processing requests from endpoints
package api

import (
	"errors"
	"strconv"

	"github.com/UnendingLoop/CommentTree/internal/model"
	"github.com/UnendingLoop/CommentTree/internal/repository"
	"github.com/UnendingLoop/CommentTree/internal/service"

	"github.com/wb-go/wbf/ginext"
)

type CommentsHandler struct {
	Service service.CommentService
}

func NewCommentHandlers(svc service.CommentService) *CommentsHandler {
	return &CommentsHandler{Service: svc}
}

func (h CommentsHandler) SimplePinger(ctx *ginext.Context) {
	ctx.JSON(200, map[string]string{"message": "pong"})
}

func (h CommentsHandler) Create(ctx *ginext.Context) {
	var newComment model.CommentCreateData

	if err := ctx.BindJSON(&newComment); err != nil {
		ctx.JSON(400, map[string]string{"error": err.Error()})
		return
	}

	res, err := h.Service.CreateComment(ctx.Request.Context(), &newComment)
	if err != nil {
		ctx.JSON(errorCodeDefiner(err), map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(201, res)
}

func (h CommentsHandler) GetAllRootComments(ctx *ginext.Context) {
	var req model.RootRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(400, map[string]string{"error": "failed to parse query"})
		return
	}

	res, err := h.Service.GetAllRootComments(ctx.Request.Context(), &req)
	if err != nil {
		ctx.JSON(errorCodeDefiner(err), map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(200, res)
}

func (h CommentsHandler) GetCommentWithChildren(ctx *ginext.Context) {
	idRaw := ctx.Param("id")
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "failed to read comment ID"})
		return
	}

	res, err := h.Service.GetCommentWithChildren(ctx.Request.Context(), id)
	if err != nil {
		ctx.JSON(errorCodeDefiner(err), map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(200, res)
}

func (h CommentsHandler) DeleteComment(ctx *ginext.Context) {
	idRaw := ctx.Param("id")
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		ctx.JSON(400, map[string]string{"error": "failed to read comment ID"})
		return
	}
	var isSoftDelete bool
	mode := ctx.Query("mode")
	switch mode {
	case "soft":
		isSoftDelete = true
	case "hard":
		isSoftDelete = false
	default:
		ctx.JSON(400, map[string]string{"error": "invalid deletion mode specified"})
		return
	}

	if err := h.Service.DeleteCommentByID(ctx.Request.Context(), id, isSoftDelete); err != nil {
		ctx.JSON(errorCodeDefiner(err), map[string]string{"error": err.Error()})
		return
	}

	ctx.Status(204)
}

func (h CommentsHandler) RunSearch(ctx *ginext.Context) {
	query := ctx.Query("q")
	if query == "" {
		ctx.JSON(400, map[string]string{"error": "empty search query"})
		return
	}

	res, err := h.Service.RunCommentSearchQuery(ctx.Request.Context(), query)
	if err != nil {
		ctx.JSON(errorCodeDefiner(err), map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(200, res)
}

func errorCodeDefiner(err error) int {
	switch {
	case errors.Is(err, service.ErrCommon500):
		return 500
	case errors.Is(err, service.ErrIncorrectQuery):
		return 400
	case errors.Is(err, service.ErrParentNotFound):
		return 404
	case errors.Is(err, service.ErrParentDeleted):
		return 409
	case errors.Is(err, service.ErrIncorrectID):
		return 400
	case errors.Is(err, repository.ErrCommentNotFound):
		return 404
	}

	return 500
}
