alter table teacher add column username text not null;
alter table teacher add column password text not null;

alter table teacher drop column clerk_id;
