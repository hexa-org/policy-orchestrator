create table integrations (
    id uuid not null primary key default gen_random_uuid(),
    name varchar(255),
    provider varchar(255),
    key bytea,
    created_at timestamp default now()
)
