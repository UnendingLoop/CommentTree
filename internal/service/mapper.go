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

func convertToDBComment(ac *APPComment) *model.DBComment {
	return &model.DBComment{
		ID:        ac.ID,
		ParentID:  ac.ParentID,
		Text:      ac.Text,
		CreatedAt: ac.CreatedAt,
		Author:    ac.Author,
	}
}

func compileToAPPCommentTree(comments []model.DBComment, rootOnly bool, parentID int) []APPComment {
	index := map[int]*APPComment{}

	// конвертируем в респонсный вид
	for _, c := range comments {
		resp := convertToAPPComment(&c)
		index[resp.ID] = resp
	}

	// связываем все узлы между собой
	for _, resp := range index {
		if resp.ParentID != nil {
			parent := index[*resp.ParentID]
			parent.Children = append(parent.Children, *resp)
		}
	}
	// сбор в итоговый массив
	result := make([]APPComment, 0)
	for _, resp := range index {
		switch rootOnly {
		case true:
			if resp.ParentID == nil {
				result = append(result, *resp)
			}
		default:
			if parentID == *resp.ParentID {
				result = append(result, *resp)
			}
		}
	}

	return result
}

func bulkConvertToAPPComments(input []model.DBComment) []APPComment {
	res := make([]APPComment, 0, len(input))
	for _, resp := range input {
		res = append(res, *convertToAPPComment(&resp))
	}
	return res
}
