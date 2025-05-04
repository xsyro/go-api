-- name: GetTodo :one
select * from todos
where id = $1 limit 1;

-- name: GetTodos :many
select * from todos;

-- name: CreateTodo :one
insert into todos (
  user_id, task
) values (
  $1, $2
)
returning *;

-- name: UpdateTodo :exec
update todos
set task = $2,
    done  = $3
where id = $1;

-- name: DeleteTodo :exec
delete from todos
where id = $1;