drop table if exists users;
create table if not exists users
(
    id        serial primary key,
    public_id varchar(100) not null unique,
    access_token varchar(100) not null,
    role      varchar(100) not null
);