-- 创建新的用户会话
-- name: CreateUserSession :exec
INSERT INTO user_sessions ( --这里的 id 其实就是 session id
    id, user_id, start_time, last_heartbeat, is_active
) VALUES (
    $1, $2, $3, $4, $5
);


