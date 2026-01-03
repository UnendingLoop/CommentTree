package service

import (
	"testing"
	"time"

	"github.com/UnendingLoop/CommentTree/internal/model"
)

func ptr[T any](v T) *T { return &v }

func TestCompileToAPPCommentTree(t *testing.T) {
	// компиляция одной ветки комментариев
	deleted := time.Now().UTC()
	comments := []model.DBComment{
		{ID: 1, Text: "root1"},
		{ID: 2, Text: "child1", ParentID: ptr(1)},
		{ID: 3, Text: "child2", ParentID: ptr(1)},
		{ID: 4, Text: "child3", ParentID: ptr(1), DeletedAt: &deleted},
	}

	tree := compileToAPPCommentTree(comments, &comments[0].ID)

	if len(tree) != 1 {
		t.Fatalf("Branch-compile: expected 1 parent, got %d", len(tree))
	}

	if len(tree[0].Children) != 3 {
		t.Fatalf("Branch-compile: expected 3 children, got %d", len(tree[0].Children))
	}

	if tree[0].Children[2].Text != deletedComment {
		t.Fatalf("Branch-compile: expected text swap for soft-deleted comment to %q, got %q", deletedComment, tree[0].Children[2].Text)
	}

	if tree[0].Children[2].CanReply != false {
		t.Fatalf("Branch-compile: expected comment to become non-replyable(false), got %v", tree[0].Children[2].CanReply)
	}

	// компиляция только корневых комментариев
	comments = []model.DBComment{
		{ID: 1, Text: "root1"},
		{ID: 2, Text: "child1", ParentID: ptr(1)},
		{ID: 3, Text: "child2", ParentID: ptr(1)},
		{ID: 4, Text: "root2"},
		{ID: 5, Text: "child1", ParentID: ptr(4)},
		{ID: 6, Text: "child2", ParentID: ptr(4)},
		{ID: 7, Text: "root3"},
		{ID: 8, Text: "child1", ParentID: ptr(7)},
		{ID: 9, Text: "child2", ParentID: ptr(7)},
	}
	tree = compileToAPPCommentTree(comments, nil)

	if len(tree) != 3 {
		t.Fatalf("Root-compile: expected 3 roots, got %d", len(tree))
	}
}

func TestConvertSearchResults(t *testing.T) {
	deleted := time.Now().UTC()
	comments := []model.DBComment{
		{ID: 1, Text: "root1"},
		{ID: 2, Text: "child1", ParentID: ptr(1)},
		{ID: 3, Text: "child2", ParentID: ptr(1)},
		{ID: 4, Text: "child3", ParentID: ptr(1), DeletedAt: &deleted},
	}

	tree := convertSearchResults(comments)

	if len(tree) != 4 {
		t.Fatalf("Search-conversion: expected 4 comments, got %d", len(tree))
	}

	if tree[3].Text != deletedComment {
		t.Fatalf("Search-conversion: expected text swap for soft-deleted comment to %q, got %q", deletedComment, tree[3].Text)
	}

	if tree[3].CanReply != false {
		t.Fatalf("Search-conversion: expected comment to become non-replyable(false), got %v", tree[3].CanReply)
	}
}
