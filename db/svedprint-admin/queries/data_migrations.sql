-- name: CreateDataMigration :one
insert into data_migrations (
    migration_name,
    description,
    executed_by,
    status
) values (
    @migration_name,
    @description,
    @executed_by,
    'pending'
) returning *;

-- name: UpdateDataMigrationStatus :exec
update data_migrations
set
    status = @status,
    error_message = @error_message,
    records_processed = @records_processed,
    records_failed = @records_failed
where id = @id;

-- name: GetDataMigrationByName :one
select * from data_migrations
where migration_name = @migration_name;

-- name: ListDataMigrations :many
select * from data_migrations
order by executed_at desc;

-- name: GetDataMigrationsByStatus :many
select * from data_migrations
where status = @status
order by executed_at desc;
