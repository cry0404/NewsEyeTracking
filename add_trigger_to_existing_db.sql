-- 为现有数据库添加comment_count自动更新触发器
-- 创建时间: 2025-07-25

-- ============================================================================
-- 1. 创建计算评论数量的函数（支持嵌套回复）
-- ============================================================================

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

-- ============================================================================
-- 2. 创建触发器函数
-- ============================================================================

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

-- ============================================================================
-- 3. 删除已存在的触发器（如果有）并创建新触发器
-- ============================================================================

-- 删除已存在的触发器（如果有）
DROP TRIGGER IF EXISTS trigger_update_comment_count ON feed_items;

-- 创建新的触发器
CREATE TRIGGER trigger_update_comment_count
    BEFORE INSERT OR UPDATE OF comments ON feed_items
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_count_trigger();

-- ============================================================================
-- 4. 输出确认信息
-- ============================================================================

SELECT 'Comment count trigger created successfully!' as status;
SELECT 'Function: calculate_comment_count() - calculates nested comment count' as function_info;
SELECT 'Trigger: trigger_update_comment_count - auto-updates comment_count on comments change' as trigger_info;
