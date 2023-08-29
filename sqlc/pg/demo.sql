create table if not exists authors
(
    id   bigserial primary key,
    name varchar not null,
    bio  text
);


-- name: StatAuthor :one
select count(*) size
     , min(id)  min_id
     , max(id)  max_id
from authors
;

-- name: ListAuthors :many
select *
from authors
order by name
;

-- name: GetAuthor :one
select *
from authors
where id = $1
limit 1
;

-- name: CreateAuthor :one
insert into authors (name, bio)
values ($1, $2)
returning *
;

-- name: DeleteAuthor :exec
delete
from authors
where id = $1
;
