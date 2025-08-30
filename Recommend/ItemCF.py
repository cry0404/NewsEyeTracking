import random
import math
import time
import pickle
import redis
from operator import itemgetter
import psycopg
from datetime import datetime, timedelta


class ItemBasedCFWithEye:
    def __init__(self):
        self.n_sim_news = 200
        self.n_rec_news = 10

        self.redis_conn = redis.Redis(host='redis', port=6379, db=1)

        self.trainSet = {}
        self.news_keywords = {}
        self.user_eye_keywords = {}
        self.news_sim_matrix = {}
        self.news_popular = {}
        self.news_count = 0

        print(f'Similar news number = {self.n_sim_news}')
        print(f'Recommended news number = {self.n_rec_news}')

    def _is_placeholder_article(self, news_id):
        if isinstance(news_id, str) and news_id.endswith('000'):
            return True
        return False

    def load_dataset_from_postgres(self, db_config):
        with psycopg.connect(**db_config) as conn:
            with conn.cursor() as cur:
                cur.execute("""
                    SELECT user_id, article_id
                    FROM reading_sessions
                """)
                read_rows = cur.fetchall()

                news_ids = set()
                for user, news in read_rows:
                    if self._is_placeholder_article(news):
                        continue
                    self.trainSet.setdefault(user, {})[news] = 1
                    news_ids.add(news)

                cur.execute("""
                    SELECT guid, keywords, published_at, created_at
                    FROM feed_items
                    WHERE guid IS NOT NULL
                """)
                keyword_rows = cur.fetchall()

                self.news_timestamps = {}

                for news_guid, news_kw_raw, published_at, created_at in keyword_rows:
                    if news_kw_raw:
                        if isinstance(news_kw_raw, list):
                            news_kw = set(kw.strip() for kw in news_kw_raw)
                        else:
                            news_kw = set(kw.strip() for kw in news_kw_raw.strip().split(','))
                    else:
                        news_kw = set()
                    self.news_keywords[news_guid] = news_kw
                    article_time = published_at if published_at else created_at
                    self.news_timestamps[news_guid] = article_time

        print(f"[INFO] Loaded {len(self.trainSet)} users and {len(self.news_keywords)} news with keywords.")

    def calc_news_sim(self):
        self.news_popular = {}
        for user, newses in self.trainSet.items():
            for news in newses:
                self.news_popular.setdefault(news, set()).add(user)
        for news in self.news_popular:
            self.news_popular[news] = len(self.news_popular[news])
        self.news_count = len(self.news_popular)
        print(f'Total news count = {self.news_count}')
        co_occurrence = {}
        for user, newses in self.trainSet.items():
            for n1 in newses:
                for n2 in newses:
                    if n1 == n2:
                        continue
                    co_occurrence.setdefault(n1, {})
                    co_occurrence[n1].setdefault(n2, 0)
                    co_occurrence[n1][n2] += 1
        print("Build co-rated matrix success")

        def keyword_similarity(n1, n2):
            kw1 = self.news_keywords.get(n1, set())
            kw2 = self.news_keywords.get(n2, set())
            if not kw1 or not kw2:
                return 0
            inter = len(kw1 & kw2)
            union = len(kw1 | kw2)
            return inter / union if union > 0 else 0

        self.news_sim_matrix = {}
        all_news_ids = set(self.news_keywords.keys())
        for n1 in all_news_ids:
            for n2 in all_news_ids:
                if n1 == n2:
                    continue
                co_count = co_occurrence.get(n1, {}).get(n2, 0)
                if self.news_popular.get(n1, 0) == 0 or self.news_popular.get(n2, 0) == 0:
                    co_sim = 0
                else:
                    co_sim = co_count / math.sqrt(self.news_popular.get(n1, 1) * self.news_popular.get(n2, 1))
                kw_sim = keyword_similarity(n1, n2)
                if co_sim > 0 or kw_sim > 0:
                    combined_sim = co_sim + kw_sim
                    self.news_sim_matrix.setdefault(n1, {})[n2] = combined_sim
                else:
                    self.news_sim_matrix.setdefault(n1, {})[n2] = 0
        print("News similarity matrix calculated.")

    def _save_pool(self, user_id, pool_data):
        self.redis_conn.set(
            f"rec_pool:{user_id}",
            pickle.dumps(pool_data),
            ex=24 * 3600
        )

    def _load_pool(self, user_id):
        data = self.redis_conn.get(f"rec_pool:{user_id}")
        return pickle.loads(data) if data else None

    def _save_recommended_history(self, user_id, recommended_news_ids):
        key = f"rec_history:{user_id}"
        existing_data = self.redis_conn.get(key)
        if existing_data:
            history = pickle.loads(existing_data)
        else:
            history = set()
        history.update(recommended_news_ids)
        self.redis_conn.set(key, pickle.dumps(history), ex=2 * 3600)

    def _get_recommended_history(self, user_id):
        key = f"rec_history:{user_id}"
        data = self.redis_conn.get(key)
        return pickle.loads(data) if data else set()

    def _filter_recent_news(self, news_list, days=2, fallback_days=7):
        """过滤最近的新闻，如果结果不足会自动扩展时间范围"""
        cutoff_time = datetime.now() - timedelta(days=days)
        filtered_news = []
        
        for item in news_list:
            if isinstance(item, tuple):
                news_id, score = item
            else:
                news_id, score = item, 0
            if self._is_placeholder_article(news_id):
                continue
            article_time = self.news_timestamps.get(news_id)
            if article_time:
                if article_time.tzinfo is not None:
                    cutoff_time = cutoff_time.replace(tzinfo=article_time.tzinfo)
                if article_time >= cutoff_time:
                    filtered_news.append((news_id, score))
            else:
                filtered_news.append((news_id, score))
        
        # 如果结果太少，扩展时间范围
        if len(filtered_news) < 5 and fallback_days > days:
            print(f"[INFO] 最近{days}天文章不足({len(filtered_news)}篇)，扩展到{fallback_days}天")
            return self._filter_recent_news(news_list, fallback_days, fallback_days * 2)
        
        return filtered_news

    def _get_recent_user_history(self, user, all_watched_news, recent_days):
        cutoff_time = datetime.now() - timedelta(days=recent_days)
        recent_watched = {}
        for news_id, rating in all_watched_news.items():
            article_time = self.news_timestamps.get(news_id)
            if article_time:
                if article_time.tzinfo is not None:
                    cutoff_time = cutoff_time.replace(tzinfo=article_time.tzinfo)
                if article_time >= cutoff_time:
                    recent_watched[news_id] = rating
            else:
                recent_watched[news_id] = rating
        return recent_watched

    # 推荐主流程参数和解包修正，变量统一
    def recommend(self, user, method='itemcf', pool_size=100):
        pool_data = self._load_pool(user)
        recs = []

        if pool_data and "pool" in pool_data and pool_data["pool"]:
            ptr = pool_data.get("ptr", 0)
            pool = pool_data["pool"]
            recs = pool[ptr:ptr + self.n_rec_news]
            pool_data["ptr"] = ptr + len(recs)
            self._save_pool(user, pool_data)

            if len(recs) < self.n_rec_news:
                new_pool, method = self._generate_recommendation_pool(user, pool_size)
                need = self.n_rec_news - len(recs)
                recs += new_pool[:need]
                self._save_pool(user, {"pool": new_pool, "ptr": need})
        else:
            new_pool, method = self._generate_recommendation_pool(user, pool_size)
            recs = new_pool[:self.n_rec_news]
            self._save_pool(user, {"pool": new_pool, "ptr": len(recs)})

        self._save_recommended_history(user, [news_id for news_id, _ in recs])
        return recs, method

    # 推荐池生成参数/变量修正
    def _generate_recommendation_pool(self, user, pool_size=100):
        recommended_history = self._get_recommended_history(user)
        all_watched_news = self.trainSet.get(user, {})
        method = 'itemcf'
        pool = []
        
        # 1. 尝试个性化推荐（基于用户历史）
        if all_watched_news:  # 用户有历史记录
            pool = self._get_personalized_recommendations(
                user, recommended_history, all_watched_news, recent_days=7, n=pool_size
            )
            print(f"[INFO] 用户{user}个性化推荐获得{len(pool)}篇文章")
        
        exist_ids = set(news_id for news_id, _ in pool) | set(all_watched_news) | set(recommended_history)
        
        # 2. 补充热门推荐
        if len(pool) < pool_size:
            need_count = pool_size - len(pool)
            hot_recs = self._get_hot_recommendations(
                user, recommended_history, all_watched_news, n=need_count, exclude_ids=exist_ids
            )
            pool.extend(hot_recs)
            exist_ids |= set(news_id for news_id, _ in hot_recs)
            print(f"[INFO] 用户{user}热门推荐补充{len(hot_recs)}篇文章")
            if not all_watched_news:  # 新用户主要依靠热门推荐
                method = 'hot'

        # 3. 补充最新文章推荐
        if len(pool) < pool_size:
            need_count = pool_size - len(pool)
            latest_recs = self._get_latest_recommendations(
                user, recommended_history, all_watched_news, n=need_count, exclude_ids=exist_ids
            )
            pool.extend(latest_recs)
            exist_ids |= set(news_id for news_id, _ in latest_recs)
            print(f"[INFO] 用户{user}最新文章补充{len(latest_recs)}篇文章")

        # 4. 最后的随机推荐兜底
        if len(pool) < pool_size:
            need_count = pool_size - len(pool)
            random_recs = self._get_random_recommendations(
                user, recommended_history, all_watched_news, n=need_count, exclude_ids=exist_ids
            )
            pool.extend(random_recs)
            print(f"[INFO] 用户{user}随机推荐兜底{len(random_recs)}篇文章")
            if len(pool) < pool_size // 2:  # 如果大部分都是随机推荐
                method = 'random'

        final_pool = pool[:pool_size]
        print(f"[INFO] 用户{user}最终推荐池大小: {len(final_pool)}, 方法: {method}")
        return final_pool, method

    def _get_personalized_recommendations(self, user, recommended_history, all_watched_news, recent_days, n=100,
                                          exclude_ids=None):
        K = self.n_sim_news
        rank = {}
        exclude_ids = exclude_ids or set()
        recent_watched_news = self._get_recent_user_history(user, all_watched_news, recent_days)
        for news, rating in recent_watched_news.items():
            for related_news, sim in sorted(self.news_sim_matrix.get(news, {}).items(), key=lambda x: x[1],
                                            reverse=True)[:K]:
                if related_news in all_watched_news or related_news in recommended_history or related_news in exclude_ids:
                    continue
                rank[related_news] = rank.get(related_news, 0) + sim
        all_recs = sorted(rank.items(), key=lambda x: x[1], reverse=True)
        filtered_recs = self._filter_recent_news(all_recs, days=3)
        return filtered_recs[:n]

    def _get_hot_recommendations(self, user, recommended_history, all_watched_news, n=100, exclude_ids=None):
        exclude_ids = exclude_ids or set()
        hot_candidates = []
        for news_id, popularity in sorted(self.news_popular.items(), key=lambda x: x[1], reverse=True):
            if news_id not in all_watched_news and news_id not in recommended_history and news_id not in exclude_ids:
                hot_candidates.append((news_id, popularity))
        
        # 先尝试7天内的热门文章
        filtered_hot = self._filter_recent_news(hot_candidates, days=7, fallback_days=14)
        
        # 如果仍然不够，不进行时间过滤，直接返回最热门的
        if len(filtered_hot) < n and len(hot_candidates) > len(filtered_hot):
            print(f"[INFO] 热门文章时间过滤后不足，使用全部热门文章")
            return hot_candidates[:n]
        
        return filtered_hot[:n]

    def _get_random_recommendations(self, user, recommended_history, all_watched_news, n=100, exclude_ids=None):
        exclude_ids = exclude_ids or set()
        all_candidates = []
        for news_id in self.news_keywords.keys():
            if news_id not in all_watched_news and news_id not in recommended_history and news_id not in exclude_ids:
                all_candidates.append((news_id, 0))
        
        # 先尝试14天内的文章
        filtered_candidates = self._filter_recent_news(all_candidates, days=14, fallback_days=30)
        
        # 如果仍然不够，使用所有可用文章
        if len(filtered_candidates) < n and len(all_candidates) > len(filtered_candidates):
            print(f"[INFO] 随机推荐时间过滤后不足，使用所有可用文章")
            random.shuffle(all_candidates)
            return all_candidates[:n]
        
        random.shuffle(filtered_candidates)
        return filtered_candidates[:n]

    def _get_latest_recommendations(self, user, recommended_history, all_watched_news, n=100, exclude_ids=None):
        exclude_ids = exclude_ids or set()
        latest_candidates = []
        
        # 收集所有符合条件的文章
        for news_id, timestamp in self.news_timestamps.items():
            if news_id not in all_watched_news and news_id not in recommended_history and news_id not in exclude_ids:
                if self._is_placeholder_article(news_id):
                    continue
                # 使用当前时间作为默认时间戳，确保没有时间戳的文章也能被推荐
                actual_timestamp = timestamp if timestamp else datetime.now()
                latest_candidates.append((news_id, actual_timestamp))
        
        # 按时间排序，最新的在前
        latest_candidates.sort(key=lambda x: x[1], reverse=True)
        latest_recs = [(news_id, 0) for news_id, _ in latest_candidates]
        
        print(f"[INFO] 最新文章推荐候选: {len(latest_recs)}篇")
        return latest_recs[:n]
