drop table if exists audit_logs;
create table if not exists audit_logs
(
    id            serial primary key,
    account_id    int FK(accounts),
    public_id     varchar(100) not null unique, --txn id
    event_id      varchar(100) not null unique,
    debit         int,
    credit        int,
    task_id       varchar(100),
    billing_cycle varchar(100)
);

drop table if exists accounts;
create table if not exists accounts
(
    id        serial primary key,
    public_id varchar(100) not null unique, --txn id
    user_id   varchar(100) not null unique,
    amount    int
);