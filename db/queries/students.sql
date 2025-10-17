-- name: GetStudentByUuid :one
select * from student
where uuid = @student_uuid;
