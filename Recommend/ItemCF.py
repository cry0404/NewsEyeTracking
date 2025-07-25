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
        self.n_sim_news = 20
        self.n_rec_news = 10

        # 初始化Redis连接（唯一新增的变量）
        self.redis_conn = redis.Redis(host='localhost', port=6379, db=1)

        self.trainSet = {}
        self.news_keywords = {}
        self.user_eye_keywords = {}
        self.news_sim_matrix = {}
        self.news_popular = {}
        self.news_count = 0

        print(f'Similar news number = {self.n_sim_news}')
        print(f'Recommended news number = {self.n_rec_news}')
    
    def _is_placeholder_article(self, news_id):
        """
        检查是否是占位符文章（列表页占位符）
        占位符文章的格式通常是：news20250724000
        """
        if isinstance(news_id, str) and news_id.endswith('000'):
            return True
        return False

    def load_dataset_from_postgres(self, db_config):
        with psycopg.connect(**db_config) as conn:
            with conn.cursor() as cur:
                # 获取阅读记录，使用 article_id (即 guid)
                cur.execute("""
                    SELECT user_id, article_id
                    FROM reading_sessions
                """)
                read_rows = cur.fetchall()

                news_ids = set()
                for user, news in read_rows:
                    # 跳过占位符文章，防止它们进入训练集
                    if self._is_placeholder_article(news):
                        continue
                    self.trainSet.setdefault(user, {})[news] = 1
                    news_ids.add(news)

                # 使用 guid 字段来匹配 reading_sessions 中的 article_id
                # 同时获取时间信息用于过滤
                cur.execute("""
                    SELECT guid, keywords, published_at, created_at
                    FROM feed_items
                    WHERE guid IS NOT NULL
                """)
                keyword_rows = cur.fetchall()

                # 存储文章时间信息
                self.news_timestamps = {}
                
                for news_guid, news_kw_raw, published_at, created_at in keyword_rows:
                    if news_kw_raw:
                        if isinstance(news_kw_raw, list):
                            news_kw = set(kw.strip() for kw in news_kw_raw)
                        else:
                            news_kw = set(kw.strip() for kw in news_kw_raw.strip().split(','))
                    else:
                        news_kw = set()

                    # 使用 guid 作为键，这样就能与 reading_sessions 中的 article_id 匹配
                    self.news_keywords[news_guid] = news_kw
                    
                    # 存储文章时间信息，优先使用 published_at，如果为空则使用 created_at
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

        # 构建共现矩阵
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
        """存储推荐池到Redis（新增方法）"""
        self.redis_conn.set(
            f"rec_pool:{user_id}",
            pickle.dumps(pool_data),
            ex=24 * 3600  # 24小时过期
        )

    def _load_pool(self, user_id):
        data = self.redis_conn.get(f"rec_pool:{user_id}")
        return pickle.loads(data) if data else None
    
    def _save_recommended_history(self, user_id, recommended_news_ids):
        """保存用户已推荐的文章历史"""
        key = f"rec_history:{user_id}"
        # 获取现有历史
        existing_data = self.redis_conn.get(key)
        if existing_data:
            history = pickle.loads(existing_data)
        else:
            history = set()
        
        # 添加新推荐的文章ID
        history.update(recommended_news_ids)
        # 保存回Redis，设置2小时过期（缩短缓存时间）
        self.redis_conn.set(key, pickle.dumps(history), ex=2 * 3600)
        
    def _get_recommended_history(self, user_id):
        """获取用户已推荐的文章历史"""
        key = f"rec_history:{user_id}"
        data = self.redis_conn.get(key)
        return pickle.loads(data) if data else set()
    
    def _filter_recent_news(self, news_list, days=2):
        """
        过滤出最近N天内的文章，同时排除占位符文章
        :param news_list: [(news_id, score), ...] 或 [news_id, ...]
        :param days: 天数限制，默认2天
        :return: 过滤后的数据， 这里是去除列表页的 news2025722000 等的影响
        """
        cutoff_time = datetime.now() - timedelta(days=days)
        
        filtered_news = []
        for item in news_list:
            # 检查是否是 (news_id, score) 对还是单纯的 news_id
            if isinstance(item, tuple):
                news_id, score = item
            else:
                news_id, score = item, 0
            
            # 首先排除占位符文章
            if self._is_placeholder_article(news_id):
                continue
            
            # 获取文章时间
            article_time = self.news_timestamps.get(news_id)
            if article_time:
                # 如果文章时间是aware的，将cutoff_time转换为同样的时区
                if article_time.tzinfo is not None:
                    cutoff_time = cutoff_time.replace(tzinfo=article_time.tzinfo)
                
                # 检查文章是否在时间范围内
                if article_time >= cutoff_time:
                    filtered_news.append((news_id, score))
            # 如果没有时间信息，默认不过滤(保留)
            else:
                filtered_news.append((news_id, score))
        
        return filtered_news

    def _get_recent_user_history(self, user, all_watched_news, recent_days):
        """
        获取用户最近N天内浏览的文章
        :param user: 用户ID
        :param all_watched_news: 用户所有浏览历史 {news_id: rating}
        :param recent_days: 最近天数
        :return: 最近浏览的文章字典 {news_id: rating}
        """
        cutoff_time = datetime.now() - timedelta(days=recent_days)
        recent_watched = {}
        
        for news_id, rating in all_watched_news.items():
            article_time = self.news_timestamps.get(news_id)
            if article_time:
                # 处理时区问题
                if article_time.tzinfo is not None:
                    cutoff_time = cutoff_time.replace(tzinfo=article_time.tzinfo)
                
                # 如果文章在最近N天内，加入到最近浏览列表
                if article_time >= cutoff_time:
                    recent_watched[news_id] = rating
            else:
                # 如果没有时间信息，默认认为是最近的（保守策略）
                recent_watched[news_id] = rating
        
        return recent_watched

    def recommend(self, user, alpha=0.0, recent_days=7):
        """
        推荐策略：
        1. 协同过滤个性化推荐
        2. 热门推荐
        3. 随机推荐
        4. 最新推荐（保底）
        """
        # 初始化推荐池和历史记录
        recommended_history = self._get_recommended_history(user)
        all_watched_news = self.trainSet.get(user, {})

        # 协同过滤推荐
        personalized_recs = self._get_personalized_recommendations(user, recommended_history, all_watched_news, recent_days)

        # 如果不够，用热门和随机推荐填充
        if len(personalized_recs) < self.n_rec_news:
            hot_recs = self._get_hot_recommendations(user, recommended_history, all_watched_news)
            personalized_recs.extend(hot_recs)

        if len(personalized_recs) < self.n_rec_news:
            random_recs = self._get_random_recommendations(user, recommended_history, all_watched_news)
            personalized_recs.extend(random_recs)

        # 添加最新推荐作为保底
        if len(personalized_recs) < self.n_rec_news:
            latest_recs = self._get_latest_recommendations(user, recommended_history, all_watched_news)
            personalized_recs.extend(latest_recs)

        # 仅返回所需数量的推荐
        final_recs = personalized_recs[:self.n_rec_news]

        # 随机打乱部分推荐，确保多样性
        stable_count = max(1, int(len(final_recs) * 0.7))
        stable_part = final_recs[:stable_count]
        random.shuffle(final_recs[stable_count:])
        final_recs = stable_part + final_recs[stable_count:]

        # 更新推荐历史
        self._save_recommended_history(user, [news_id for news_id, _ in final_recs])

        return final_recs

    def _get_personalized_recommendations(self, user, recommended_history, all_watched_news, recent_days):
        """
        协同过滤个性化推荐
        """
        K = self.n_sim_news
        rank = {}
        
        # 只使用用户最近N天内浏览的文章作为推荐种子
        recent_watched_news = self._get_recent_user_history(user, all_watched_news, recent_days)
        
        # 基于最近浏览的文章进行推荐
        for news, rating in recent_watched_news.items():
            for related_news, sim in sorted(self.news_sim_matrix.get(news, {}).items(),
                                            key=itemgetter(1), reverse=True)[:K]:
                # 避免推荐用户已经看过的文章
                if related_news in all_watched_news:
                    continue
                # 避免推荐用户已经收到过的文章
                if related_news in recommended_history:
                    continue
                    
                rank[related_news] = rank.get(related_news, 0) + sim

        # 对排序后的推荐结果进行时间过滤
        all_recs = sorted(rank.items(), key=itemgetter(1), reverse=True)
        filtered_recs = self._filter_recent_news(all_recs, days=3)
        
        return filtered_recs[:self.n_rec_news]

    def _get_hot_recommendations(self, user, recommended_history, all_watched_news):
        """
        基于流行度的热门推荐
        """
        # 按流行度排序，排除已看过和已推荐的
        hot_candidates = []
        for news_id, popularity in sorted(self.news_popular.items(), key=itemgetter(1), reverse=True):
            if news_id not in all_watched_news and news_id not in recommended_history:
                hot_candidates.append((news_id, popularity))
        
        # 时间过滤
        filtered_hot = self._filter_recent_news(hot_candidates, days=3)
        return filtered_hot[:self.n_rec_news]

    def _get_random_recommendations(self, user, recommended_history, all_watched_news):
        """
        随机推荐
        """
        # 从所有文章中过滤出可推荐的
        all_candidates = []
        for news_id in self.news_keywords.keys():
            if news_id not in all_watched_news and news_id not in recommended_history:
                all_candidates.append((news_id, 0))
        
        # 时间过滤
        filtered_candidates = self._filter_recent_news(all_candidates, days=3)
        
        # 随机选择
        if len(filtered_candidates) <= self.n_rec_news:
            return filtered_candidates
        
        return random.sample(filtered_candidates, self.n_rec_news)

    def _get_latest_recommendations(self, user, recommended_history, all_watched_news):
        """
        最新文章推荐（保底策略）
        """
        # 按时间排序，获取最新的文章
        latest_candidates = []
        for news_id, timestamp in self.news_timestamps.items():
            if news_id not in all_watched_news and news_id not in recommended_history:
                if timestamp:  # 确保有时间戳
                    latest_candidates.append((news_id, timestamp))
        
        # 按时间降序排序
        latest_candidates.sort(key=lambda x: x[1], reverse=True)
        
        # 转换为 (news_id, score) 格式
        latest_recs = [(news_id, 0) for news_id, _ in latest_candidates[:self.n_rec_news]]
        
        return latest_recs

    def random_recommend(self, N=None, user_id=None):
        if N is None:
            N = self.n_rec_news
            
        # 从所有文章中过滤出2天内的文章
        all_news_with_scores = [(news, 0) for news in self.news_popular.keys()]
        recent_news = self._filter_recent_news(all_news_with_scores, days=2)
        
        # 如果指定了user_id，还要过滤掉已推荐和已观看的文章
        if user_id:
            recommended_history = self._get_recommended_history(user_id)
            all_watched_news = self.trainSet.get(user_id, {})
            
            filtered_news = []
            for news_id, score in recent_news:
                if news_id not in recommended_history and news_id not in all_watched_news:
                    filtered_news.append((news_id, score))
            recent_news = filtered_news
        
        if len(recent_news) <= N:
            return recent_news
        
        # 从过滤后的文章中随机选取
        sampled = random.sample(recent_news, N)
        return sampled
