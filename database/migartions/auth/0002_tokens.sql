drop table if exists tokens;
create table if not exists tokens
(
    public_id    varchar(100) not null unique,
    access_token varchar(100) not null
);