drop table if exists tasks;
create table if not exists tasks
(
    id          serial primary key,
    public_id   varchar(100) not null unique,
    title       varchar(256) not null,
    jira        varchar(256),
    description text         not null,
    assignee_id varchar(100) not null,
    status      varchar(100) not null
);