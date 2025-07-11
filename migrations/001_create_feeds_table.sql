-- +goose Up
-- 创建feeds表 - 存储RSS源信息
CREATE TABLE feeds (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    link VARCHAR(512) NOT NULL,
    rsshub_route VARCHAR(255) NOT NULL UNIQUE, -- rsshub路由地址
    feed_url VARCHAR(512) NOT NULL, -- 完整的RSS URL
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建feed_items表 - 存储文章内容
CREATE TABLE feed_items (
    id SERIAL PRIMARY KEY,
    feed_id INTEGER REFERENCES feeds(id) ON DELETE CASCADE, --有关的文章内容
    title VARCHAR(512) NOT NULL,
    description TEXT,
    content TEXT, -- 文章完整内容
    link VARCHAR(512) NOT NULL,
    guid VARCHAR(512) NOT NULL UNIQUE, -- 文章唯一标识
    author VARCHAR(255),
    keywords VARCHAR(255) ARRAY,
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 创建索引以提升查询性能
CREATE INDEX idx_feeds_rsshub_route ON feeds(rsshub_route);
CREATE INDEX idx_feeds_is_active ON feeds(is_active);
CREATE INDEX idx_feed_items_feed_id ON feed_items(feed_id);
CREATE INDEX idx_feed_items_published_at ON feed_items(published_at);
CREATE INDEX idx_feed_items_guid ON feed_items(guid);

-- +goose Down
DROP TABLE IF EXISTS feed_items;
DROP TABLE IF EXISTS feeds;