-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE events
    ADD COLUMN visitor_hash VARCHAR(64),
    ADD COLUMN country_code VARCHAR(2),
    ADD COLUMN country_name VARCHAR(100),
    ADD COLUMN city VARCHAR(100),
    ADD COLUMN region VARCHAR(100);

CREATE INDEX idx_events_visitor ON events(website_id, visitor_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS idx_events_visitor;
ALTER TABLE events
    DROP COLUMN IF EXISTS visitor_hash,
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS country_name,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS region;
-- +goose StatementEnd
