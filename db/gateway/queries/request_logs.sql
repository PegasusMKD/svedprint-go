-- name: InsertRequestLog :exec
insert into request_logs (
    timestamp,
    method,
    incoming_path,
    redirected_path,
    user_id,
    realm,
    status_code,
    response_time_ms,
    upstream_service,
    error_message
) values (
    @timestamp,
    @method,
    @incoming_path,
    @redirected_path,
    @user_id,
    @realm,
    @status_code,
    @response_time_ms,
    @upstream_service,
    @error_message
);

-- name: BatchInsertRequestLogs :copyfrom
insert into request_logs (
    timestamp,
    method,
    incoming_path,
    redirected_path,
    user_id,
    realm,
    status_code,
    response_time_ms,
    upstream_service,
    error_message
) values (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetRequestLogsByUser :many
select * from request_logs
where user_id = @user_id
order by timestamp desc
limit @limit_count;

-- name: GetRequestLogsByTimeRange :many
select * from request_logs
where timestamp between @start_time and @end_time
order by timestamp desc;

-- name: GetRecentErrorLogs :many
select * from request_logs
where status_code >= 400
order by timestamp desc
limit @limit_count;

-- name: GetRequestLogsByService :many
select * from request_logs
where upstream_service = @upstream_service
order by timestamp desc
limit @limit_count;
