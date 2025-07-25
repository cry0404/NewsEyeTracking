-- +goose Up
ALTER TABLE feed_items ADD COLUMN comment_count INT DEFAULT 0;

-- +goose Down
ALTER TABLE feed_items DROP COLUMN comment_count;