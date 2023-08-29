create table if not exists authors
(
    id   bigint       not null auto_increment primary key,
    name varchar(512) not null,
    bio  text
)
;

-- name: StatAuthor :one
select count(*) size
     , min(id)  min_id
     , max(id)  max_id
from authors
;

-- name: ListAuthors :many
select *
from authors
order by name;


-- name: GetAuthor :one
select *
from authors
where id = ?
    limit 1;

-- name: CreateAuthor :execresult
insert into authors (name, bio)
values (?, ?);

-- name: DeleteAuthor :exec
delete
from authors
where id = ?;
