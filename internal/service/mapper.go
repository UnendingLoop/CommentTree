package service

import (
	"time"

	"commentTree/internal/model"
)

type APPComment struct {
	ID        int          `json:"id,omitempty"`
	ParentID  *int         `json:"parent_id,omitempty"`
	Text      string       `json:"content"`
	CreatedAt time.Time    `json:"created_at,omitempty"`
	IsDeleted bool         `json:"deleted,omitempty"`
	CanReply  bool         `json:"replyable,omitempty"`
	Author    string       `json:"author,omitempty"`
	Children  []APPComment `json:"children,omitempty"`
}

func convertToAPPComment(c *model.DBComment) *APPComment {
	isDeleted := c.DeletedAt != nil

	content := c.Text
	if isDeleted {
		content = "[Комментарий удалён]"
	}

	return &APPComment{
		ID:        c.ID,
		ParentID:  c.ParentID,
		Text:      content,
		CreatedAt: c.CreatedAt,
		IsDeleted: isDeleted,
		CanReply:  !isDeleted,
	}
}

func compileToAPPCommentTree(comments []model.DBComment, parentID *int) []APPComment {
	index := map[int]*APPComment{}

	// конвертируем в респонсный вид
	for _, c := range comments {
		resp := convertToAPPComment(&c)
		index[resp.ID] = resp
	}

	// связываем все узлы между собой
	for _, resp := range index {
		if resp.ParentID != nil {
			parent, ok := index[*resp.ParentID]
			if ok {
				parent.Children = append(parent.Children, *resp)
			}
		}
	}
	// сбор в итоговый массив
	result := make([]APPComment, 0)
	for _, resp := range index {
		switch parentID {
		case nil: // ищем только корни
			if resp.ParentID == nil {
				result = append(result, *resp)
			}
		default:
			if resp.ID == *parentID { // ищем только родителя всей ветки
				result = append(result, *resp)
				return result
			}
		}
	}

	return result
}

func convertSearchResults(input []model.DBComment) []APPComment {
	res := make([]APPComment, 0, len(input))
	for _, resp := range input {
		res = append(res, *convertToAPPComment(&resp))
	}
	return res
}
