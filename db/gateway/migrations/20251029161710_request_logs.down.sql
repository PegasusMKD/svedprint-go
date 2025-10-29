-- Drop gateway database tables

drop index if exists idx_request_logs_service;
drop index if exists idx_request_logs_errors;
drop index if exists idx_request_logs_user_id;
drop index if exists idx_request_logs_timestamp;

drop table if exists request_logs;
