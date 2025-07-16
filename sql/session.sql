-- 创建一个新的会话，相当于打开了一篇新的网页，开启了新的事件
-- 初始的 endtime 应该是为空的, oss 存储路径应该暂定, 这里的 starttime 应该是有的
-- 但是 endtime 应该是待定的，由前端发回来的信息， sessionid 作为打包区分，
-- 一个 user_id 的内容

-- name: CreateSession :one
INSERT INTO reading_sessions (
    id, article_id, start_time, device_info, user_id
)VALUES(
    uuid_generate_v4(), $1, $2, $3, $4
) RETURNING *;

-- 会话查询
-- name: GetSessionByID :one
SELECT * FROM reading_sessions WHERE id = $1;

-- name: GetUserActiveSessions :many
SELECT * FROM reading_sessions WHERE user_id = $1 AND end_time IS NULL;

-- 会话更新操作
-- name: UpdateSessionEndTime :one
UPDATE reading_sessions SET
    end_time = $2,
    session_duration_ms = EXTRACT(EPOCH FROM ($2 - start_time)) * 1000
WHERE id = $1 RETURNING *;

-- 会话统计查询
-- name: GetUserSessionStats :many
SELECT 
    article_id,
    COUNT(*) as session_count,
    AVG(session_duration_ms) as avg_duration_ms,
    SUM(event_count) as total_events
FROM reading_sessions 
WHERE user_id = $1 AND end_time IS NOT NULL
GROUP BY article_id;
