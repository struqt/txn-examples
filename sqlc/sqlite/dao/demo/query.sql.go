// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: query.sql

package demo

import (
	"context"
	"database/sql"
)

const createAuthor = `-- name: CreateAuthor :one
;

insert into authors (name, bio)
values (?, ?) returning id, name, bio
`

type CreateAuthorParams struct {
	Name string
	Bio  sql.NullString
}

func (q *Queries) CreateAuthor(ctx context.Context, arg CreateAuthorParams) (Author, error) {
	row := q.queryRow(ctx, q.createAuthorStmt, createAuthor, arg.Name, arg.Bio)
	var i Author
	err := row.Scan(&i.ID, &i.Name, &i.Bio)
	return i, err
}

const deleteAuthor = `-- name: DeleteAuthor :exec
;

delete
from authors
where id = ?
`

func (q *Queries) DeleteAuthor(ctx context.Context, id int64) error {
	_, err := q.exec(ctx, q.deleteAuthorStmt, deleteAuthor, id)
	return err
}

const getAuthor = `-- name: GetAuthor :one
;

select id, name, bio
from authors
where id = ? limit 1
`

func (q *Queries) GetAuthor(ctx context.Context, id int64) (Author, error) {
	row := q.queryRow(ctx, q.getAuthorStmt, getAuthor, id)
	var i Author
	err := row.Scan(&i.ID, &i.Name, &i.Bio)
	return i, err
}

const listAuthors = `-- name: ListAuthors :many
;

select id, name, bio
from authors
order by name
`

func (q *Queries) ListAuthors(ctx context.Context) ([]Author, error) {
	rows, err := q.query(ctx, q.listAuthorsStmt, listAuthors)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Author
	for rows.Next() {
		var i Author
		if err := rows.Scan(&i.ID, &i.Name, &i.Bio); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const statAuthor = `-- name: StatAuthor :one
select count(*) size
     , min(id)  min_id
     , max(id)  max_id
from authors
`

type StatAuthorRow struct {
	Count int64
	Min   interface{}
	Max   interface{}
}

func (q *Queries) StatAuthor(ctx context.Context) (StatAuthorRow, error) {
	row := q.queryRow(ctx, q.statAuthorStmt, statAuthor)
	var i StatAuthorRow
	err := row.Scan(&i.Count, &i.Min, &i.Max)
	return i, err
}