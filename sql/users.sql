-- 用户 CRUD 操作
-- name: CreateUser :one
-- 用户 ID 来自邀请码 ID，建立一对一关系
INSERT INTO users (
    id, email, gender, age, date_of_birth, education_level, residence,
    weekly_reading_hours, primary_news_platform, is_active_searcher,
    is_colorblind, vision_status, is_vision_corrected
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET
    gender = $2, age = $3, date_of_birth = $4, education_level = $5, residence = $6,
    weekly_reading_hours = $7, primary_news_platform = $8, is_active_searcher = $9,
    is_colorblind = $10, vision_status = $11, is_vision_corrected = $12
WHERE id = $1 RETURNING *;

-- A/B 测试相关查询
-- name: GetUserWithInviteCode :one
SELECT u.*, ic.has_recommend, ic.has_more_information 
FROM users u 
JOIN invite_codes ic ON u.id = ic.id 
WHERE u.id = $1;

-- name: GetUserABTestConfig :one
SELECT has_recommend, has_more_information 
FROM invite_codes WHERE id = $1;
