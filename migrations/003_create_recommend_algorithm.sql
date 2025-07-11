-- +goose Up

-- 用户点击文章事件记录表
CREATE TABLE user_click_events (
    id SERIAL PRIMARY KEY,                                    -- 独立主键，允许多次点击
    user_id UUID NOT NULL,                                   -- 用户ID
    article_id INTEGER NOT NULL,                             -- 文章ID（feed_items表的id）
    session_id UUID,                                         -- 关联阅读会话（可选）
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),       -- 点击时间
    ip_address INET,                                         -- 点击IP（可选，用于分析）
    user_agent TEXT,                                         -- 用户代理（可选，用于设备分析）
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES feed_items(id) ON DELETE CASCADE,
    FOREIGN KEY (session_id) REFERENCES reading_sessions(id) ON DELETE SET NULL
);

-- 创建索引优化查询性能
CREATE INDEX idx_user_click_events_user_id ON user_click_events(user_id);
CREATE INDEX idx_user_click_events_article_id ON user_click_events(article_id);
CREATE INDEX idx_user_click_events_clicked_at ON user_click_events(clicked_at);
CREATE INDEX idx_user_click_events_user_time ON user_click_events(user_id, clicked_at);

-- +goose Down

DROP TABLE user_click_events;