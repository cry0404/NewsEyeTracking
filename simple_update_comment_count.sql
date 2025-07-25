-- 简化版：更新comment_count字段
-- 仅包含必要的更新逻辑，不创建持久函数或视图

-- ============================================================================
-- 方案1：使用临时函数更新comment_count（推荐）
-- ============================================================================

-- 创建临时函数计算评论数量（包含嵌套回复）
CREATE OR REPLACE FUNCTION temp_count_comments(comments_json JSONB)
RETURNS INTEGER AS $$
DECLARE
    comment_record JSONB;
    total_count INTEGER := 0;
    replies_count INTEGER := 0;
BEGIN
    -- 如果没有评论，返回0
    IF comments_json IS NULL OR jsonb_array_length(comments_json) = 0 THEN
        RETURN 0;
    END IF;
    
    -- 遍历所有评论
    FOR comment_record IN SELECT * FROM jsonb_array_elements(comments_json)
    LOOP
        -- 计算当前评论
        total_count := total_count + 1;
        
        -- 递归计算回复
        IF comment_record ? 'replies' AND 
           comment_record->'replies' IS NOT NULL AND
           jsonb_array_length(comment_record->'replies') > 0 THEN
            replies_count := temp_count_comments(comment_record->'replies');
            total_count := total_count + replies_count;
        END IF;
    END LOOP;
    
    RETURN total_count;
END;
$$ LANGUAGE plpgsql;

-- 更新所有文章的comment_count
UPDATE feed_items 
SET comment_count = temp_count_comments(comments)
WHERE comments IS NOT NULL;

-- 清理临时函数
DROP FUNCTION temp_count_comments(JSONB);

-- ============================================================================
-- 方案2：简单更新（如果评论结构是平面数组）
-- ============================================================================

-- 如果comments只是简单的数组结构，不包含嵌套回复，可以使用：
-- UPDATE feed_items 
-- SET comment_count = COALESCE(jsonb_array_length(comments), 0)
-- WHERE comments IS NOT NULL;

-- ============================================================================
-- 验证更新结果
-- ============================================================================

-- 查看更新统计
SELECT 
    COUNT(*) as total_articles,
    COUNT(CASE WHEN comments IS NOT NULL THEN 1 END) as articles_with_comments,
    COUNT(CASE WHEN comment_count > 0 THEN 1 END) as articles_with_positive_count,
    SUM(comment_count) as total_comment_count,
    MAX(comment_count) as max_comments
FROM feed_items;

-- 查看有评论的文章示例
SELECT id, LEFT(title, 60) as title, comment_count, created_at
FROM feed_items 
WHERE comment_count > 0
ORDER BY comment_count DESC
LIMIT 5;
