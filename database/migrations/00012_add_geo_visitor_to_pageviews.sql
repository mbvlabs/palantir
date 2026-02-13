-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE pageviews
    ADD COLUMN visitor_hash VARCHAR(64),
    ADD COLUMN country_code VARCHAR(2),
    ADD COLUMN country_name VARCHAR(100),
    ADD COLUMN city VARCHAR(100),
    ADD COLUMN region VARCHAR(100);

CREATE INDEX idx_pageviews_visitor ON pageviews(website_id, visitor_hash);
CREATE INDEX idx_pageviews_country ON pageviews(website_id, country_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_pageviews_country;
DROP INDEX IF EXISTS idx_pageviews_visitor;
ALTER TABLE pageviews
    DROP COLUMN IF EXISTS visitor_hash,
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS country_name,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS region;
-- +goose StatementEnd
