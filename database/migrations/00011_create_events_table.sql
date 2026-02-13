-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS events (
    id uuid not null PRIMARY KEY,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,

    website_id uuid NOT NULL REFERENCES websites(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    event_name VARCHAR(255) NOT NULL,
    event_data JSONB
);

CREATE INDEX idx_events_website_created ON events(website_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
