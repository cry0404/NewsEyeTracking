-- 获取新的文章， 这里需要根据推荐算法，所以这里筛选出来的接口还应该需要接到推荐算法上
-- name: GetNewArticles :many
SELECT * FROM feed_items
WHERE published_at > $1  -- 这里每天根据推荐时间选取
ORDER BY published_at DESC
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