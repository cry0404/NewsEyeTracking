import psycopg

def get_user_strategy(db_config, user_id):
    sql = "SELECT has_recommend FROM invite_codes WHERE id = %s"
    with psycopg.connect(**db_config) as conn:
        with conn.cursor() as cur:
            cur.execute(sql, (user_id,))
            result = cur.fetchone()
            if result is None:
                return 'random'  # 默认策略
            has_recommend = result[0]
            return 'itemcf' if has_recommend else 'random'
