create type gender as enum ('male', 'female', 'unknown');
create type study_type as enum ('regular', 'irregular');
create type academic_level as enum ('first_year', 'second_year', 'junior_year', 'senior_year');
create type behaviour_type as enum ('exemplary'); -- TODO: Add more??
create type year_success_type as enum ('excellent', 'good', 'fair', 'unsatisfactory', 'fail');

create table student (
	uuid uuid primary key,
	first_name text,
	middle_name text,
	last_name text,
	personal_number text,
	fathers_name text,
	mothers_name text,
	date_of_birth date,
	place_of_residence text,
	place_of_birth text,
	citizenship text,
	school_uuid uuid not null
);

create table students_yearly_detail (
	uuid uuid primary key,
	student_uuid uuid not null references student (uuid),
	academic_level academic_level not null,

	number_in_class int,
	
	taken_exam bool not null default false,
	passed_exam bool not null default false,
	passed_year bool not null default false,
	justified_absences int not null default 0,
	unjustified_absences int not null default 0,
	behaviour_type behaviour_type not null default 'exemplary',
	primted_testimony bool not null default false,
	testimony_print_date date,
	business_number date,
	diploma_business_number text,
	initial_subject_package_uuid uuid not null,

	school_class_uuid uuid not null,
	academic_year_uuid uuid not null,

	constraint uq_student_yearly_detail unique (student_uuid, academic_level, academic_year_uuid, school_class_uuid)
);

create table students_yearly_optional_subjects (
	uuid uuid primary key,
	students_yearly_detail uuid not null references students_yearly_detail (uuid),
	subject_uuid uuid not null
);

