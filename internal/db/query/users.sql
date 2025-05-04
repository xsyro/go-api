-- name: GetUserByID :one
select * from users
where id = $1;

-- name: GetUserByEmail :one
select * from users 
where email = $1;

-- name: GetUsers :many
select * from users;

-- name: CreateUser :one
insert into users (
    name,uh
    email, 
    password, 
    role,
    password_changed_at, 
    created_at, 
    account_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) returning *;

-- name: UpdateUserByID :exec
update users set 
    name = $2,
    email = $3,
    password = $4,
    role = $5,
    password_changed_at = $6,
    account_id = $7
where id = $1;

-- name: DeleteUserByID :exec
delete from users where id = $1;

-- name: SetUserPassword :exec
update users set
    password = $2,
    password_changed_at = now()
where id = $1;

-- name: SetUserRole :exec
update users set
    role = $2
where id = $1;
