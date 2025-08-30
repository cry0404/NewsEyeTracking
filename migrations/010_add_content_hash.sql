-- +goose Up

-- 添加 content_hash 字段，用于内容去重
ALTER TABLE feed_items ADD COLUMN content_hash VARCHAR(64);

-- 为 content_hash 字段创建索引以提高查询性能
CREATE INDEX idx_feed_items_content_hash ON feed_items(content_hash);

-- 为现有数据生成 content_hash（如果有数据的话）
-- UPDATE feed_items SET content_hash = md5(COALESCE(title, '') || COALESCE(link, '') || COALESCE(description, ''))
-- WHERE content_hash IS NULL;

-- +goose Down

-- 删除索引和字段
DROP INDEX IF EXISTS idx_feed_items_content_hash;
ALTER TABLE feed_items DROP COLUMN content_hash;
