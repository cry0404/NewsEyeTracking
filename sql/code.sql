-- 获取邀请码ID 和 email（注册时使用）
-- name: GetIdAndEmailByCode :one
SELECT id,email FROM invite_codes WHERE code = $1;


-- 验证邀请码并自动增加使用次数计数, 这里就算没注册也应该算使用了
-- name: FindCodeAndIncrementCount :one
-- 如果 code 存在，就增加 count，无论 is_used 是什么
UPDATE invite_codes
SET count = COALESCE(count, 0) + 1
WHERE code = $1
RETURNING id, code, is_used, count;



-- 只查询邀请码信息（不增加计数，用于纯查询场景）
-- name: MarkInviteCodeAsUsed :exec
UPDATE invite_codes 
SET is_used = TRUE 
WHERE code = $1;




-- 根据邀请码ID获取A/B测试配置，但是这里也许该再解耦一下，毕竟 has_more infomation 应该是只需要查询一次的，没必要一直查询
-- name: GetABTestConfigByInviteCodeID :one
SELECT has_recommend, has_more_information 
FROM invite_codes 
WHERE id = $1;

-- 但是以后登录都是使用邀请码，所以登录时应该发放 jwt？