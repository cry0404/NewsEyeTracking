-- name: CreateUser :one
-- 这里的 id 应该由前端默认填写，根据邀请码的信息来查表
INSERT INTO users (
    id, email, gender, age, date_of_birth, education_level, residence,
    weekly_reading_hours, primary_news_platform, is_active_searcher,
    is_colorblind, vision_status, is_vision_corrected
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;
-- 这里对应的应该是注册时的 

-- 通过传入的第一个参数来更新返回对应的 user 信息
-- name: UpdateUser :one
UPDATE users SET
    gender = $2, age = $3, date_of_birth = $4, education_level = $5, residence = $6,
    weekly_reading_hours = $7, primary_news_platform = $8, is_active_searcher = $9,
    is_colorblind = $10, vision_status = $11, is_vision_corrected = $12
WHERE id = $1 RETURNING *;

-- 根据邀请码关联用户来查询对应的 recommend 和 information 状态
-- name: GetUserABtestInformation :one
SELECT has_recommend, has_more_information 
FROM invite_codes WHERE id = $1;