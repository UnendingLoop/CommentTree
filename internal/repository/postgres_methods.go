package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"commentTree/internal/model"
)

func (p PostgresRepo) Create(ctx context.Context, n *model.CommentCreateData) (*model.DBComment, error) {
	query := `INSERT INTO comments (cid, pid, content, created_at, author)
	VALUES (DEFAULT, $1, $2, DEFAULT, $3) 
	RETURNING cid, pid, content, created_at, author`
	res := model.DBComment{}
	if err := p.db.QueryRowContext(ctx, query, n.ParentID, n.Text, n.Author).Scan(&res.ID, &res.ParentID, &res.Text, &res.CreatedAt, &res.Author); err != nil {
		return nil, err
	}
	return &res, nil
}

func (p PostgresRepo) GetCommentByID(ctx context.Context, id int) (*model.DBComment, error) {
	query := `SELECT pid, content, created_at, deleted_at, author FROM comments WHERE cid = $1`
	var comment model.DBComment
	comment.ID = id

	err := p.db.QueryRowContext(ctx, query, id).Scan(&comment.ParentID, &comment.Text, &comment.CreatedAt, &comment.DeletedAt, &comment.Author)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrCommentNotFound // 404
		default:
			return nil, err
		}
	}

	return &comment, nil
}

func (p PostgresRepo) GetAllRoot(ctx context.Context, limit, offset int, sort, order string) ([]model.DBComment, error) {
	query := fmt.Sprintf(`SELECT cid, pid, content, created_at, deleted_at, author 
	FROM comments
	ORDER BY %s %s 
	LIMIT $1 
	OFFSET $2`, sort, order)

	rows, err := p.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := make([]model.DBComment, 0, limit)
	for rows.Next() {
		var c model.DBComment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt, &c.DeletedAt, &c.Author); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return comments, nil
}

func (p PostgresRepo) GetCommentWithChildrenByID(ctx context.Context, id int) ([]model.DBComment, error) {
	query := `WITH RECURSIVE comment_tree AS (
    SELECT *
    FROM comments
    WHERE cid = $1

    UNION ALL

    SELECT c.*
    FROM comments c
    JOIN comment_tree ct ON c.pid = ct.cid
    WHERE c.deleted_at IS NULL
	)

	SELECT cid, pid, content, created_at, deleted_at, author 
	FROM comment_tree
	ORDER BY pid ASC, created_at DESC`

	rows, err := p.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := make([]model.DBComment, 0)
	for rows.Next() {
		var c model.DBComment
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt, &c.DeletedAt, &c.Author); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return comments, nil
}

func (p PostgresRepo) DeleteByID(ctx context.Context, id int) error {
	query := `WITH RECURSIVE comment_tree AS (
    SELECT *
    FROM comments
    WHERE cid = $1

    UNION ALL

    SELECT c.*
    FROM comments c
    JOIN comment_tree ct ON c.pid = ct.cid
    WHERE c.deleted_at IS NULL
	)

	DELETE FROM comments
	WHERE cid IN (
    SELECT cid FROM comment_tree
	)`

	row := p.db.QueryRowContext(ctx, query, id)
	if row.Err() != nil {
		switch {
		case errors.Is(row.Err(), sql.ErrNoRows):
			return ErrCommentNotFound // 404
		default:
			return row.Err()
		}
	}
	return nil
}

func (p PostgresRepo) MarkAsDeletedByID(ctx context.Context, id int) error {
	query := `UPDATE comments
	SET deleted_at = $1 WHERE cid = $2`

	row := p.db.QueryRowContext(ctx, query, time.Now().UTC(), id)
	if row.Err() != nil {
		switch {
		case errors.Is(row.Err(), sql.ErrNoRows):
			return ErrCommentNotFound // 404
		default:
			return row.Err()
		}
	}
	return nil
}

func (p PostgresRepo) RunSearchQuery(ctx context.Context, q string) ([]model.DBComment, error) {
	query := `SELECT cid, pid, content, created_at, author,
	ts_rank(content_tsv, websearch_to_tsquery('simple', $1)) AS rank
	FROM comments
	WHERE deleted_at IS NULL
	AND content_tsv @@ websearch_to_tsquery('russian', $1)
	ORDER BY rank DESC, created_at DESC;`
	rows, err := p.db.QueryContext(ctx, query, q)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := make([]model.DBComment, 0)
	for rows.Next() {
		var c model.DBComment
		rank := 0.99
		if err := rows.Scan(&c.ID, &c.ParentID, &c.Text, &c.CreatedAt, &c.Author, &rank); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return comments, nil
}
