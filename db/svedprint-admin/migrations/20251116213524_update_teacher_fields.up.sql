alter table teacher drop column username;
alter table teacher drop column password;

alter table teacher add column clerk_id text not null;
