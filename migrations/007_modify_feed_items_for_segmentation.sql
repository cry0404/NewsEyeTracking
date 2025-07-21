-- +goose Up
-- 修改feed_items表字段以支持分词处理
-- 分词后的文本会包含大量的<span>标签，需要更大的存储空间

-- 1. 修改title字段：从VARCHAR(512)改为TEXT
-- 原因：分词后标题长度可能增长20-30倍
ALTER TABLE feed_items ALTER COLUMN title TYPE TEXT;

-- 2. 修改keywords字段：从VARCHAR(255) ARRAY改为TEXT ARRAY
-- 原因：每个关键词可能包含特殊字符，需要更大空间
ALTER TABLE feed_items ALTER COLUMN keywords TYPE TEXT ARRAY;

-- 3. 为分词处理添加注释
COMMENT ON COLUMN feed_items.title IS '文章标题（支持分词标签）';
COMMENT ON COLUMN feed_items.description IS '文章描述（支持分词标签）';
COMMENT ON COLUMN feed_items.content IS '文章内容（支持分词标签）';
COMMENT ON COLUMN feed_items.keywords IS '关键词数组（通过分词算法提取）';

-- +goose Down
-- 回滚：将字段改回原来的类型
-- 注意：如果数据已经超过原有长度限制，回滚可能失败
ALTER TABLE feed_items ALTER COLUMN title TYPE VARCHAR(512);
ALTER TABLE feed_items ALTER COLUMN keywords TYPE VARCHAR(255) ARRAY;

-- 移除注释
COMMENT ON COLUMN feed_items.title IS NULL;
COMMENT ON COLUMN feed_items.description IS NULL;  
COMMENT ON COLUMN feed_items.content IS NULL;
COMMENT ON COLUMN feed_items.keywords IS NULL;
