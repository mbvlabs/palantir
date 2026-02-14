-- name: InsertPageview :one
insert into
    pageviews (id, created_at, website_id, url, referrer, browser, os, device, country, language, screen_width, visitor_hash, country_code, country_name, city, region)
values
    ($1, now(), $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
returning *;

-- name: QueryPageviewsPerDay :many
select date_trunc('day', created_at)::timestamptz as date, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by date order by date;

-- name: QueryTotalPageviews :one
select count(*)::bigint as total
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz;

-- name: QueryTopPages :many
select url, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by url order by views desc limit 10;

-- name: QueryTopReferrers :many
select referrer, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
  and referrer is not null and referrer != ''
group by referrer order by views desc limit 10;

-- name: QueryBrowserBreakdown :many
select browser, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by browser order by views desc;

-- name: QueryOSBreakdown :many
select os, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by os order by views desc;

-- name: QueryDeviceBreakdown :many
select device, count(*)::bigint as views
from pageviews
where website_id = $1 and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by device order by views desc;

-- name: QueryPageviewsTimeBucketed :many
select date_trunc(sqlc.arg('bucket')::text, created_at)::timestamptz as bucket_time,
       count(*)::bigint as views
from pageviews
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
group by bucket_time order by bucket_time;

-- name: QueryUniqueVisitorsTimeBucketed :many
select date_trunc(sqlc.arg('bucket')::text, created_at)::timestamptz as bucket_time,
       count(distinct visitor_hash)::bigint as visitors
from pageviews
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
  and visitor_hash is not null
group by bucket_time order by bucket_time;

-- name: QueryTotalUniqueVisitors :one
select count(distinct visitor_hash)::bigint as total
from pageviews
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
  and visitor_hash is not null;

-- name: QueryTopCountries :many
select country_code, country_name, count(*)::bigint as views
from pageviews
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
  and country_code is not null and country_code != ''
group by country_code, country_name order by views desc limit 10;

-- name: QueryTopCities :many
select city, country_code, count(*)::bigint as views
from pageviews
where website_id = $1
  and created_at between sqlc.arg('start_date')::timestamptz and sqlc.arg('end_date')::timestamptz
  and city is not null and city != ''
group by city, country_code order by views desc limit 10;

-- name: QueryBounceCount :one
SELECT count(*)::bigint AS bounce_visitors
FROM (
    SELECT visitor_hash
    FROM pageviews
    WHERE website_id = $1
      AND created_at BETWEEN sqlc.arg('start_date')::timestamptz AND sqlc.arg('end_date')::timestamptz
      AND visitor_hash IS NOT NULL
    GROUP BY visitor_hash
    HAVING count(*) = 1
) AS single_page_visitors;
