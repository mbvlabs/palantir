-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS pageviews (
    id uuid not null PRIMARY KEY,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,

    website_id uuid NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    referrer TEXT,
    browser VARCHAR(64),
    os VARCHAR(64),
    device VARCHAR(32),
    country VARCHAR(2),
    language VARCHAR(16),
    screen_width INT
);

CREATE INDEX idx_pageviews_website_created ON pageviews(website_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS pageviews;
-- +goose StatementEnd
