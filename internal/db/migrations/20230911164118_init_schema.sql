-- +goose Up
-- +goose StatementBegin
create extension if not exists  "uuid-ossp";

create type role as enum ('guest', 'member', 'admin');

create table accounts (
	id uuid primary key default uuid_generate_v4(),
	email text unique not null
);

create table users (
	id uuid primary key default uuid_generate_v4(),
	name text not null,
	email text unique not null,
	password text not null,
	role role not null default 'guest',
	password_changed_at timestamptz,
	created_at timestamptz not null default (now()),
	account_id uuid not null references accounts(id) on delete cascade
);

create table todos (
    id uuid primary key default uuid_generate_v4(),
    task text not null,
    done boolean not null default false,
	created_at timestamptz not null default (now()),
    user_id uuid not null references users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
drop table accounts;
drop table todos;
drop extension "uuid-ossp";
-- +goose StatementEnd
