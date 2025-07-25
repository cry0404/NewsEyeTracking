-- PostgreSQL数据库初始化脚本
-- 整合了所有迁移文件，用于生产环境部署
-- 创建时间: 2025-01-24
-- 版本: v1.0

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- 1. 创建feeds表 - 存储RSS源信息
-- ============================================================================
CREATE TABLE feeds (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    link VARCHAR(512) NOT NULL,
    rsshub_route VARCHAR(255) NOT NULL UNIQUE, -- rsshub路由地址
    feed_url VARCHAR(512) NOT NULL, -- 完整的RSS URL
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- 2. 创建feed_items表 - 存储文章内容
-- ============================================================================
CREATE TABLE feed_items (
    id SERIAL PRIMARY KEY,
    feed_id INTEGER REFERENCES feeds(id) ON DELETE CASCADE, -- 关联的RSS源
    title TEXT NOT NULL, -- 文章标题（支持分词标签）
    description TEXT, -- 文章描述（支持分词标签）
    content TEXT, -- 文章完整内容（支持分词标签）
    link VARCHAR(512) NOT NULL,
    guid VARCHAR(512) NOT NULL UNIQUE, -- 文章唯一标识
    author VARCHAR(255),
    keywords TEXT ARRAY, -- 关键词数组（通过分词算法提取）
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    -- 新增统计字段
    like_count INT DEFAULT 0,
    share_count INT DEFAULT 0,
    save_count INT DEFAULT 0,
    comment_count INT DEFAULT 0,
    comments JSONB -- 存储嵌套的评论结构
);

-- 为分词处理添加注释
COMMENT ON COLUMN feed_items.title IS '文章标题（支持分词标签）';
COMMENT ON COLUMN feed_items.description IS '文章描述（支持分词标签）';
COMMENT ON COLUMN feed_items.content IS '文章内容（支持分词标签）';
COMMENT ON COLUMN feed_items.keywords IS '关键词数组（通过分词算法提取）';

-- ============================================================================
-- 3. 创建邀请码表 - A/B测试配置
-- ============================================================================
CREATE TABLE invite_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    code VARCHAR(50) NOT NULL UNIQUE,
    is_used BOOLEAN DEFAULT FALSE,
    has_recommend BOOLEAN DEFAULT FALSE,  -- A/B测试：是否启用推荐算法
    has_more_information BOOLEAN DEFAULT FALSE,  -- A/B测试：是否显示更多信息
    count INTEGER DEFAULT 0, -- 邀请码使用次数
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- 4. 创建用户表 - 用户基本信息和阅读习惯
-- ============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY, -- id 来自于 invite_codes
    email VARCHAR(255) NOT NULL UNIQUE,
    gender VARCHAR(10),
    age INT,
    date_of_birth DATE,
    education_level VARCHAR(50),
    residence VARCHAR(100),
    
    -- 新闻阅读习惯
    weekly_reading_hours INT,
    primary_news_platform VARCHAR(100),
    is_active_searcher BOOLEAN DEFAULT FALSE,
    
    -- 视觉相关
    is_colorblind BOOLEAN DEFAULT FALSE,
    vision_status VARCHAR(50),  -- 正常/近视/远视
    is_vision_corrected BOOLEAN DEFAULT FALSE,
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- 外键约束，确立与 invite_codes 的一对一关系
    FOREIGN KEY (id) REFERENCES invite_codes(id) ON DELETE CASCADE
);

-- ============================================================================
-- 5. 创建用户登录会话表 - 管理用户登录状态
-- ============================================================================
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- session_id
    user_id UUID NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_heartbeat TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE,
    created_date DATE DEFAULT CURRENT_DATE,
    end_time TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    
    FOREIGN KEY (user_id) REFERENCES invite_codes(id) ON DELETE CASCADE
);

-- ============================================================================
-- 6. 创建阅读会话表 - 记录用户阅读行为
-- ============================================================================
CREATE TABLE reading_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL, -- 引用 users.id
    article_id VARCHAR(512) NOT NULL,  -- 文章ID，列表页定义为 "0"
    start_time TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    device_info JSONB,  -- 记录设备信息（如分辨率等）
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    -- 注意：不再有外键约束到 feed_items，允许更灵活的文章ID管理
);

-- ============================================================================
-- 创建索引 - 优化查询性能
-- ============================================================================

-- feeds表索引
CREATE INDEX idx_feeds_rsshub_route ON feeds(rsshub_route);
CREATE INDEX idx_feeds_is_active ON feeds(is_active);

-- feed_items表索引
CREATE INDEX idx_feed_items_feed_id ON feed_items(feed_id);
CREATE INDEX idx_feed_items_published_at ON feed_items(published_at);
CREATE INDEX idx_feed_items_guid ON feed_items(guid);

-- invite_codes表索引
CREATE INDEX idx_invite_codes_code ON invite_codes(code);

-- users表索引
CREATE INDEX idx_users_email ON users(email);

-- user_sessions表索引
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_active ON user_sessions(is_active) WHERE is_active = TRUE;
CREATE INDEX idx_user_sessions_heartbeat ON user_sessions(last_heartbeat) WHERE is_active = TRUE;
CREATE INDEX idx_user_sessions_active_heartbeat ON user_sessions(user_id, is_active, last_heartbeat);

-- 改进的用户会话索引
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_active_heartbeat 
    ON user_sessions(user_id, is_active, last_heartbeat) 
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_user_sessions_expired 
    ON user_sessions(is_active, last_heartbeat) 
    WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_user_sessions_status 
    ON user_sessions(id, is_active, last_heartbeat);

-- reading_sessions表索引
CREATE INDEX idx_sessions_user_article ON reading_sessions(user_id, article_id);
CREATE INDEX idx_sessions_start_time ON reading_sessions(start_time);
CREATE INDEX idx_sessions_user_time ON reading_sessions(user_id, start_time);
CREATE INDEX idx_sessions_end_time ON reading_sessions(end_time);

-- ============================================================================
-- 7. 创建comment_count自动更新触发器
-- ============================================================================

-- 创建计算评论数量的函数（支持嵌套回复）
CREATE OR REPLACE FUNCTION calculate_comment_count(comments_json JSONB)
RETURNS INTEGER AS $$
DECLARE
    comment_record JSONB;
    total_count INTEGER := 0;
    replies_count INTEGER := 0;
BEGIN
    -- 如果评论为空，返回0
    IF comments_json IS NULL OR jsonb_array_length(comments_json) = 0 THEN
        RETURN 0;
    END IF;
    
    -- 遍历评论数组
    FOR comment_record IN SELECT * FROM jsonb_array_elements(comments_json)
    LOOP
        -- 计算当前评论
        total_count := total_count + 1;
        
        -- 递归计算回复
        IF comment_record ? 'replies' AND 
           comment_record->'replies' IS NOT NULL AND
           jsonb_array_length(comment_record->'replies') > 0 THEN
            replies_count := calculate_comment_count(comment_record->'replies');
            total_count := total_count + replies_count;
        END IF;
    END LOOP;
    
    RETURN total_count;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器函数
CREATE OR REPLACE FUNCTION update_comment_count_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- 在INSERT或UPDATE时自动计算comment_count
    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        NEW.comment_count := calculate_comment_count(NEW.comments);
        RETURN NEW;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
CREATE TRIGGER trigger_update_comment_count
    BEFORE INSERT OR UPDATE OF comments ON feed_items
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_count_trigger();

-- ============================================================================
-- 数据库初始化完成
-- ============================================================================

-- 输出完成信息
SELECT 'NewsEyeTracking数据库初始化完成！' AS status;
SELECT 'Created tables: feeds, feed_items, invite_codes, users, user_sessions, reading_sessions' AS tables_created;
SELECT 'All indexes and constraints have been applied.' AS indexes_status;
SELECT 'Comment count trigger has been created.' AS trigger_status;
