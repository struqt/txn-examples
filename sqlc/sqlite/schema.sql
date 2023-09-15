create table if not exists authors
(
    id   integer primary key,
    name text not null,
    bio  text
);
