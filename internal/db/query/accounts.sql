-- name: GetAccountByID :one
select * from  accounts
where id = $1 limit 1;

-- name: CreateAccount :one
insert into accounts (
    email
) values (
    $1
) returning *;

-- name: UpdateAccountByID :exec
update accounts set 
    email = $2
where id = $1;

-- name: DeleteAccountByID :exec
delete from accounts
where id = $1;
