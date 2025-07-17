-- 创建新的用户会话
-- name: CreateUserSession :one
INSERT INTO user_sessions (
    user_id, start_time, last_heartbeat, is_active
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- 获取用户的活跃会话
-- name: GetUserActiveSession :one
SELECT * FROM user_sessions 
WHERE user_id = $1 
AND is_active = TRUE 
AND last_heartbeat > NOW() - INTERVAL '5 minutes'
ORDER BY start_time DESC 
LIMIT 1;

-- 更新会话心跳时间
-- name: UpdateSessionHeartbeat :exec
UPDATE user_sessions 
SET last_heartbeat = $2 
WHERE id = $1 AND is_active = TRUE;

-- 结束用户会话
-- name: EndUserSession :exec
UPDATE user_sessions 
SET is_active = FALSE 
WHERE id = $1;

-- 查找过期的会话（超过5分钟没有心跳）
-- name: FindExpiredSessions :many
SELECT * FROM user_sessions 
WHERE is_active = TRUE 
AND last_heartbeat < NOW() - INTERVAL '5 minutes';

-- 批量设置过期会话为非活跃状态
-- name: MarkExpiredSessionsInactive :exec
UPDATE user_sessions 
SET is_active = FALSE 
WHERE is_active = TRUE 
AND last_heartbeat < NOW() - INTERVAL '5 minutes';

-- 获取用户会话详情（包含相关的阅读会话）
-- name: GetUserSessionWithReadingSessions :one
SELECT 
    us.*,
    COUNT(rs.id) as reading_session_count
FROM user_sessions us
LEFT JOIN reading_sessions rs ON us.id = rs.user_session_id
WHERE us.id = $1
GROUP BY us.id, us.user_id, us.start_time, us.last_heartbeat, us.is_active, us.created_date;

-- 检查是否有其他用户的活跃会话（用于单会话限制）
-- name: CheckOtherActiveUserSessions :many
SELECT * FROM user_sessions 
WHERE user_id != $1 
AND is_active = TRUE 
AND last_heartbeat > NOW() - INTERVAL '5 minutes';

-- 强制结束其他用户的活跃会话
-- name: ForceEndOtherUserSessions :exec
UPDATE user_sessions 
SET is_active = FALSE 
WHERE user_id != $1 
AND is_active = TRUE;
