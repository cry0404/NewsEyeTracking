-- +goose Up

-- 增加 like_count, share_count, save_count 字段，并设置默认值
ALTER TABLE feed_items ADD COLUMN like_count INT DEFAULT 0;
ALTER TABLE feed_items ADD COLUMN share_count INT DEFAULT 0;
ALTER TABLE feed_items ADD COLUMN save_count INT DEFAULT 0;

-- 增加 comments 字段，使用 JSONB 类型来存储嵌套的评论结构
-- JSONB 是一种二进制格式，通常比 JSON 类型更高效
ALTER TABLE feed_items ADD COLUMN comments JSONB;

-- +goose Down
-- 移除这些字段
ALTER TABLE feed_items DROP COLUMN like_count;
ALTER TABLE feed_items DROP COLUMN share_count;
ALTER TABLE feed_items DROP COLUMN save_count;
ALTER TABLE feed_items DROP COLUMN comments;
