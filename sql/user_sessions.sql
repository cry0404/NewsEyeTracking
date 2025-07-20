-- 创建新的用户会话
-- name: CreateUserSession :exec
INSERT INTO user_sessions ( --这里的 id 其实就是 session id
    id, user_id, start_time, last_heartbeat, is_active
) VALUES (
    $1, $2, $3, $4, $5
);

-- 更新心跳时间并自动检查过期（主要使用的心跳更新方法）
-- name: UpdateHeartbeatWithExpireCheck :exec
UPDATE user_sessions 
SET 
    last_heartbeat = $1::timestamp,
    end_time = CASE 
        WHEN EXTRACT(EPOCH FROM ($1::timestamp - last_heartbeat)) > $2::integer THEN $1::timestamp
        ELSE end_time 
    END,
    is_active = CASE 
        WHEN EXTRACT(EPOCH FROM ($1::timestamp - last_heartbeat)) > $2::integer THEN FALSE
        ELSE is_active 
    END
WHERE id = $3 AND is_active = TRUE;

-- 根据会话ID获取会话信息
-- name: GetUserSessionByID :one
SELECT id, user_id, start_time, last_heartbeat, is_active, end_time, created_date
FROM user_sessions 
WHERE id = $1;

-- 根据用户ID获取活跃会话
-- name: GetActiveUserSessionByUserID :one
SELECT id, user_id, start_time, last_heartbeat, is_active, end_time, created_date
FROM user_sessions 
WHERE user_id = $1 AND is_active = TRUE 
ORDER BY last_heartbeat DESC 
LIMIT 1;

-- 手动结束会话（用于登出等场景）
-- name: EndUserSession :exec
UPDATE user_sessions 
SET 
    end_time = $1,
    is_active = FALSE
WHERE id = $2;

-- 批量清理所有过期会话（定时任务使用）
-- name: CleanupExpiredSessions :exec
UPDATE user_sessions 
SET 
    end_time = COALESCE(end_time, NOW()),
    is_active = FALSE
WHERE 
    is_active = TRUE 
    AND EXTRACT(EPOCH FROM (NOW() - last_heartbeat)) > $1::integer;

-- 获取用户的所有会话（包括已结束的）
-- name: GetUserSessionsByUserID :many
SELECT id, user_id, start_time, last_heartbeat, is_active, end_time, created_date
FROM user_sessions 
WHERE user_id = $1 
ORDER BY start_time DESC;


