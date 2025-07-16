-- +goose Up
ALTER TABLE invite_codes ADD COLUMN count INTEGER DEFAULT 0;
-- 统计邀请码的使用次数，所以对应的登录逻辑也需要改
-- +goose Down
ALTER TABLE invite_codes DROP COLUMN count;