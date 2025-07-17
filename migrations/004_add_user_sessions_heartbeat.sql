-- +goose Up
-- 创建用户登录会话表
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_date DATE DEFAULT CURRENT_DATE,
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 为reading_sessions表添加user_session_id字段
ALTER TABLE reading_sessions ADD COLUMN user_session_id UUID;

-- 添加外键约束
ALTER TABLE reading_sessions ADD CONSTRAINT fk_reading_sessions_user_session 
    FOREIGN KEY (user_session_id) REFERENCES user_sessions(id) ON DELETE CASCADE;

-- 创建索引以提高查询性能
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_user_sessions_heartbeat ON user_sessions(last_heartbeat) WHERE is_active = TRUE;
CREATE INDEX idx_user_sessions_active_heartbeat ON user_sessions(user_id, is_active, last_heartbeat);
CREATE INDEX idx_reading_sessions_user_session ON reading_sessions(user_session_id);

-- +goose Down
-- 删除索引
DROP INDEX IF EXISTS idx_reading_sessions_user_session;
DROP INDEX IF EXISTS idx_user_sessions_active_heartbeat;
DROP INDEX IF EXISTS idx_user_sessions_heartbeat;
DROP INDEX IF EXISTS idx_user_sessions_active;
DROP INDEX IF EXISTS idx_user_sessions_user_id;

-- 删除外键约束和字段
ALTER TABLE reading_sessions DROP CONSTRAINT IF EXISTS fk_reading_sessions_user_session;
ALTER TABLE reading_sessions DROP COLUMN IF EXISTS user_session_id;

-- 删除用户会话表
DROP TABLE IF EXISTS user_sessions;
