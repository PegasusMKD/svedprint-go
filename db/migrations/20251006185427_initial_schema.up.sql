create type gender as enum ('male', 'female', 'unknown');
create type study_type as enum ('regular', 'irregular');
create type academic_level as enum ('first_year', 'second_year', 'junior_year', 'senior_year');
create type behaviour_type as enum ('exemplary'); -- TODO: Add more??
create type year_success_type as enum ('excellent', 'good', 'fair', 'unsatisfactory', 'fail');
create type subject_orientations as enum ('general', 'pma', 'pmb'); -- TODO: Maybe remove this as an idea completely?
create type school_type as enum ('gymnasium', 'professional');

create table school (
	uuid uuid primary key,
	school_name text not null unique,
	director_name text,
	business_number text,
	main_book text,
	ministry text,
	country text,
	city text
);

create table academic_year (
	school_uuid uuid not null references school (uuid),
	year_range text not null,
	school_type school_type not null,
	last_digits_of_year text not null,
	act_number text,
	act_date date,
	primary key (school_uuid, school_type, year_range)
);

create table school_diploma_details (
	uuid uuid primary key,
	academic_year_uuid uuid references academic_year (uuid),
	print_date date not null
);

create table school_testimony_details (
	uuid uuid primary key,
	academic_year_uuid uuid references academic_year (uuid),
	testimony_date date
);


create table teacher (
	uuid uuid primary key,
	school_uuid uuid not null references school (uuid),
	
	first_name text not null,
	middle_name text,
	last_name text not null,

	username text not null,
	password text not null,

	print_allowed bool not null default false
);

create table subject (
	uuid uuid primary key,
	short_name text,
	full_name text,
	academic_level academic_level not null,
	school_uuid uuid not null references school (uuid)
);

create table subject_package (
	uuid uuid primary key,
	short_name text,
	full_name text,
	academic_level academic_level not null,
	academic_year_uuid uuid not null references academic_year (uuid)
);

create table subject_package_subjects (
	uuid uuid primary key,
	subject_uuid uuid references subject (uuid),
	subject_package_uuid uuid not null references subject_package (uuid)
);


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
	school_uuid uuid not null references school (uuid)
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
	initial_subject_package_uuid uuid not null references subject_package (uuid),

	school_class_uuid uuid not null references school_class (uuid),
	academic_year_uuid uuid not null references academic_year (uuid),

	constraint uq_student_yearly_detail unique (student_uuid, academic_level, academic_year_uuid, school_class_uuid)
);

create table students_yearly_optional_subjects (
	uuid uuid primary key,
	students_yearly_detail uuid not null references students_yearly_detail (uuid)
);

create table school_class (
	uuid uuid primary key,
	academic_level academic_level,
	suffix text,
	academic_year_uuid uuid references academic_year (uuid),
	responsible_teacher_uuid uuid references teacher (uuid),
	default_subject_package_uuid uuid references subject_package (uuid)
);


