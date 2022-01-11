create table integrations (
    id         uuid not null primary key default gen_random_uuid(),
    name       varchar(255),
    provider   varchar(255),
    key        bytea,
    created_at timestamp default now()
);

create table applications (
    id             uuid not null primary key default gen_random_uuid(),
    integration_id uuid,
    object_id      varchar(255),
    name           varchar(255),
    description    varchar(255),
    created_at     timestamp default now(),
    constraint fk_integration
        foreign key (integration_id)
            references integrations (id) on delete cascade
);
