-- name: StatAuthor :one
select count(*) size
     , min(id)  min_id
     , max(id)  max_id
from authors
;

-- name: GetAuthor :one
select *
from authors
where id = ? limit 1
;

-- name: ListAuthors :many
select *
from authors
order by name
;

-- name: CreateAuthor :one
insert into authors (name, bio)
values (?, ?) returning *
;

-- name: DeleteAuthor :exec
delete
from authors
where id = ?
;
