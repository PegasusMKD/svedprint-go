-- name: GetSchoolByUuid :one
select * from school
where uuid = @school_uuid;

-- name: InsertSchool :one
insert into school (uuid, school_name)
values (gen_random_uuid(), @school_name)
returning *;

-- name: UpdateSchool :one
update school
set director_name = @director_name
where uuid = @school_uuid
returning *
;
