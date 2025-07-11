-- 这里作为一开始注册的时候返回给前端的字段，让他们自动填到对应的 user 表中
-- name: GetIdByCode :one
SELECT id FROM invite_codes WHERE code = $1;