-- +goose Up
-- invite_codes 表的设计保持相对独立，只记录邀请码本身的状态。
-- 用户的创建将依赖于一个有效的、未使用的邀请码。
CREATE TABLE invite_codes (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                VARCHAR(255) NOT NULL UNIQUE,
    code                 VARCHAR(50) NOT NULL UNIQUE,
    is_used              BOOLEAN DEFAULT FALSE,
    has_recommend        BOOLEAN DEFAULT FALSE,  -- A/B测试：是否启用推荐算法
    has_more_information BOOLEAN DEFAULT FALSE,  -- A/B测试：是否显示更多信息
    created_at           TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- users 表的 id 直接引用自 invite_codes 表的 id。
-- 这样，每个用户都唯一对应一个邀请码，关系清晰。
CREATE TABLE users (
    id                    UUID PRIMARY KEY, -- id 来自于 invite_codes, 所以没有默认值
    email                 VARCHAR(255) NOT NULL UNIQUE,
    gender                VARCHAR(10),
    age                   INT,
    date_of_birth         DATE,
    education_level       VARCHAR(50),
    residence             VARCHAR(100),
    
    -- 新闻阅读习惯
    weekly_reading_hours  INT,
    primary_news_platform VARCHAR(100),
    is_active_searcher    BOOLEAN DEFAULT FALSE,
    
    -- 视觉相关
    is_colorblind         BOOLEAN DEFAULT FALSE,
    vision_status         VARCHAR(50),  -- 正常/近视/远视
    is_vision_corrected   BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at            TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 外键约束，确立与 invite_codes 的一对一关系
    FOREIGN KEY (id) REFERENCES invite_codes(id) ON DELETE CASCADE
);


-- 创建阅读会话表（使用UUID作为主键）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- reding 就代表一个会话，包括阅读主页，以及阅读文章，
-- 
CREATE TABLE reading_sessions (
    id                   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id              UUID NOT NULL, -- 引用 users.id, 类型必须是 UUID
    article_id           INTEGER NOT NULL,  -- 将列表页定义为 0 ， 这样就独立了列表的数据记录？
    start_time           TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_time             TIMESTAMP WITH TIME ZONE,
    device_info          JSONB,  -- 这里可以记录分辨率
    
    
    session_duration_ms  BIGINT,  -- 感觉其实不太需要
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (article_id) REFERENCES feed_items(id) ON DELETE CASCADE -- 确保 feed_items 表存在
);

-- 索引
CREATE INDEX idx_invite_codes_code ON invite_codes(code); -- 为邀请码查询添加索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_sessions_user_article ON reading_sessions(user_id, article_id);
CREATE INDEX idx_sessions_start_time ON reading_sessions(start_time);
CREATE INDEX idx_sessions_user_time ON reading_sessions(user_id, start_time);

CREATE INDEX idx_sessions_end_time ON reading_sessions(end_time);

-- - invite_codes: 只需要 created_at (邀请码不会被修改)
-- - users: 只需要 created_at (用户信息更新在应用层处理)
-- - reading: 使用 start_time 和 end_time 跟踪会话生命周期

-- +goose Down
DROP TABLE IF EXISTS reading_sessions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS invite_codes; 

