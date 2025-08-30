-- 获取新的文章， 这里需要根据推荐算法，所以这里筛选出来的接口还应该需要接到推荐算法上
-- name: GetNewArticles :many
SELECT * FROM feed_items
WHERE published_at > $1  -- 这里每天根据推荐时间选取
ORDER BY published_at DESC
LIMIT $2;

-- name: GetArticlesByGUID :many
SELECT id, title, content, guid, like_count, share_count, save_count, comment_count
FROM feed_items 
WHERE guid = ANY($1::text[])
ORDER BY published_at DESC
LIMIT $2;

-- 随机获取新闻文章
-- name: GetRandomArticles :many
SELECT * FROM feed_items
WHERE published_at > $1  -- 最近一段时间的新闻
ORDER BY RANDOM()
LIMIT $2;

-- 获取文章的详细信息， 这里需要根据文章的id来获取
-- name: GetArticleByID :one
SELECT feed_id, title, description, content, link, guid, author, published_at 
FROM feed_items 
WHERE id = $1;

-- name: GetArticleByGUID :one 
-- 根据 guid 来应该更好一点， 因为guid是唯一的， 而id不是
SELECT feed_id, title, description, content, link, guid, author, published_at 
FROM feed_items 
WHERE guid = $1;

-- name: GetMoreInfoMation :one
SELECT like_count, share_count, save_count, comments
FROM feed_items
WHERE guid = $1;

-- name: GetFeedItemByContentHash :one
-- 根据内容哈希查找是否存在重复文章
SELECT id, guid, title, link, created_at
FROM feed_items
WHERE content_hash = $1
LIMIT 1;

-- name: UpsertFeedItem :one
-- 插入或更新文章（现在包含content_hash字段）
INSERT INTO feed_items (
    feed_id, title, description, content, link, guid, author, 
    published_at, like_count, share_count, save_count, comments, content_hash
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
ON CONFLICT (guid) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    content = EXCLUDED.content,
    link = EXCLUDED.link,
    author = EXCLUDED.author,
    published_at = EXCLUDED.published_at,
    like_count = EXCLUDED.like_count,
    share_count = EXCLUDED.share_count,
    save_count = EXCLUDED.save_count,
    comments = EXCLUDED.comments,
    content_hash = EXCLUDED.content_hash
RETURNING *;
