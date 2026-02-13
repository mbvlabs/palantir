-- name: InsertEvent :one
insert into
    events (id, created_at, website_id, url, event_name, event_data, visitor_hash, country_code, country_name, city, region)
values
    ($1, now(), $2, $3, $4, $5, $6, $7, $8, $9, $10)
returning *;

-- name: QueryTopEvents :many
select event_name, count(*)::bigint as event_count
from events
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by event_name order by event_count desc limit 10;

-- name: QueryEventsTimeBucketed :many
select date_trunc(sqlc.arg('bucket')::text, created_at)::timestamptz as bucket_time,
       count(*)::bigint as event_count
from events
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by bucket_time order by bucket_time;
