-- +goose Up
ALTER TABLE reading_sessions DROP CONSTRAINT reading_sessions_article_id_fkey;

-- +goose Down
ALTER TABLE reading_sessions ADD CONSTRAINT reading_sessions_article_id_fkey FOREIGN KEY (article_id) REFERENCES feed_items(guid) ON DELETE CASCADE;