-- 获取邀请码ID（注册时使用）
-- name: GetIdByCode :one
SELECT id FROM invite_codes WHERE code = $1;

-- 验证邀请码（检查是否存在且未使用）
-- name: ValidateInviteCode :one
SELECT id, code, is_used, has_recommend, has_more_information, created_at 
FROM invite_codes 
WHERE code = $1 AND (is_used IS NULL OR is_used = FALSE);

-- 标记邀请码为已使用
-- name: MarkInviteCodeAsUsed :one
UPDATE invite_codes 
SET is_used = TRUE 
WHERE code = $1 
RETURNING id, code, is_used, has_recommend, has_more_information, created_at;

-- 根据邀请码ID获取A/B测试配置，但是这里也许该再解耦一下，毕竟 has_more infomation 应该是只需要查询一次的，没必要一直查询
-- name: GetABTestConfigByInviteCodeID :one
SELECT has_recommend, has_more_information 
FROM invite_codes 
WHERE id = $1;