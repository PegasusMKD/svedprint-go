create table if not exists request_logs (
    id uuid primary key default gen_random_uuid(),
    timestamp timestamptz not null default now(),
    method varchar(10) not null,
    incoming_path text not null,
    redirected_path text,
    user_id text,
    realm text,
    status_code int not null,
    response_time_ms int not null,
    upstream_service text,
    error_message text,
    created_at timestamptz not null default now()
);

create index idx_request_logs_timestamp on request_logs (timestamp desc);
create index idx_request_logs_user_id on request_logs (user_id) where user_id is not null;
create index idx_request_logs_errors on request_logs (status_code) where status_code >= 400;
create index idx_request_logs_service on request_logs (upstream_service);
