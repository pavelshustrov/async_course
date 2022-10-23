drop table if exists jobs;
create table if not exists jobs
(
    id            serial primary key,
    event_name    varchar(100) not null,
    event_version varchar(100) not null,
    payload       jsonb        not null,
    created_at    timestamp    not null default now(),
    deleted_at    timestamp
);
