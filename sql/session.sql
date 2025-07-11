-- 创建一个新的会话，相当于打开了一篇新的网页，开启了新的事件
-- 初始的 endtime 应该是为空的, oss 存储路径应该暂定, 这里的 starttime 应该是有的
-- 但是 endtime 应该是待定的，由前端发回来的信息， sessionid 作为打包区分，
-- 一个 user_id 的内容

-- name: CreateSession :one
INSERT INTO reading_sessions (
    article_id, start_time, device_info, user_id
)VALUES(
    $1, $2, $3, $4
) RETURNING *;

-- 更新会话的结束时间， 这里应该根据 sessionid 来更新,前端发送过来
-- name: UpdateSessionEndTime :one
UPDATE reading_sessions SET
    end_time = $2
WHERE id = $1 RETURNING *;
