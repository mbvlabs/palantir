-- name: QueryWebsiteByID :one
select * from websites where id=$1;

-- name: QueryWebsitesByUserID :many
select * from websites where user_id=$1 order by created_at desc;

-- name: InsertWebsite :one
insert into
    websites (id, created_at, updated_at, user_id, name, domain)
values
    ($1, now(), now(), $2, $3, $4)
returning *;

-- name: UpdateWebsite :one
update websites
    set updated_at=now(), name=$2, domain=$3
where id = $1
returning *;

-- name: DeleteWebsite :exec
delete from websites where id=$1;
