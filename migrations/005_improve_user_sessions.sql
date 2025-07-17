-- +goose Up
-- 改进用户会话管理，移除基于日期的限制
-- 添加更好的索引以支持基于心跳时间的会话查询

-- 添加复合索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_active_heartbeat 
    ON user_sessions(user_id, is_active, last_heartbeat) 
    WHERE is_active = TRUE;

-- 添加过期会话清理的索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_expired 
    ON user_sessions(is_active, last_heartbeat) 
    WHERE is_active = TRUE;

-- 添加会话状态查询的索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_status 
    ON user_sessions(id, is_active, last_heartbeat);

-- +goose Down
-- 删除添加的索引
DROP INDEX IF EXISTS idx_user_sessions_status;
DROP INDEX IF EXISTS idx_user_sessions_expired;  
DROP INDEX IF EXISTS idx_user_sessions_user_active_heartbeat;
