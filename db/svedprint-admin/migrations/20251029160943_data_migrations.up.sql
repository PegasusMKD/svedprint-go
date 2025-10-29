create type migration_status as enum ('pending', 'in_progress', 'completed', 'failed', 'rolled_back');

-- Data migration tracking
create table data_migrations (
    id bigserial primary key,
    migration_name varchar(255) not null unique,
    description text,
    executed_by uuid not null,
    executed_at timestamptz not null default now(),
    status migration_status not null,
    error_message text,
    records_processed int default 0,
    records_failed int default 0
);

-- Indexes
create index idx_data_migrations_status on data_migrations (status);

