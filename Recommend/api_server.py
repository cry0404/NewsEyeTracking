from flask import Flask, request, jsonify
import psycopg
from ItemCF import ItemBasedCFWithEye  # 你保存上面类的模块名

app = Flask(__name__)

db_config = {
    'host': 'localhost',
    'port': 5432,
    'user': 'cry',
    'password': '',
    'dbname': 'easyrss'
}

# 全局变量存储推荐器
recommender = None

def initialize_recommender():
    """初始化推荐系统"""
    global recommender
    try:
        print("开始初始化推荐系统...")
        recommender = ItemBasedCFWithEye()
        recommender.load_dataset_from_postgres(db_config)
        recommender.calc_news_sim()
        print("推荐系统初始化完成。")
        return True
    except Exception as e:
        print(f"推荐系统初始化失败: {e}")
        import traceback
        traceback.print_exc()
        return False

def get_user_strategy(user_id):
    # 通过用户UUID从invite_codes表中查询推荐策略
    sql = "SELECT has_recommend FROM invite_codes WHERE id = %s"
    with psycopg.connect(**db_config) as conn:
        with conn.cursor() as cur:
            cur.execute(sql, (user_id,))
            result = cur.fetchone()
            if result is None:
                return 'random'
            return 'itemcf' if result[0] else 'random'
@app.route('/')
def index():
    return jsonify({"message": "推荐系统运行中，请使用 POST /recommend 访问推荐接口"})

@app.route('/recommend', methods=['POST'])
def recommend():
    try:
        data = request.get_json()
        user_id = data.get("user_id")
        if not user_id:
            return jsonify({"error": "缺少 user_id 参数"}), 400

        # 支持UUID字符串作为用户ID
        # 如果用户不存在于训练集中，使用随机推荐
        strategy = get_user_strategy(user_id)
        
        # 确保user_id是UUID类型
        if isinstance(user_id, str):
            import uuid
            try:
                user_uuid = uuid.UUID(user_id)
            except ValueError:
                user_uuid = user_id
        else:
            user_uuid = user_id
            
        if user_uuid in recommender.trainSet and strategy == 'itemcf':
            recs = recommender.recommend(user_uuid)
        else:
            recs = recommender.random_recommend(user_id=user_uuid)
            
        return jsonify({
            "user_id": user_id,
            "strategy": strategy,
            "recommendations": [
                {"news_id": news_id, "score": round(score, 4)}
                for news_id, score in recs
            ]
        })
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    # 初始化推荐系统
    if not initialize_recommender():
        print("推荐系统初始化失败，退出...")
        exit(1)
    
    app.run(host='0.0.0.0', port=6667, debug=False)  # 关闭debug模式避免重启
