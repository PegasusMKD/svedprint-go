-- name: InsertTeacher :one
insert into teacher (school_uuid, first_name, middle_name, last_name, clerk_id)
values (@school_uuid, @first_name, @middle_name, @last_name, @clerk_id)
returning *;
