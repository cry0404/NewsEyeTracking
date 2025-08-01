from flask import Flask, request, jsonify
import jieba
import psycopg
from bs4 import BeautifulSoup, NavigableString
from textcut import HtmlTextSegmenter
import os

print("系统默认编码:", os.sys.getdefaultencoding())
print(os.path.abspath(__file__))

app = Flask(__name__)
segmenter = HtmlTextSegmenter()

db_config = {
    'host': 'postgres',
    'port': 5432,
    'user': 'cry',
    'password': '',
    'dbname': 'easyrss'
}

@app.route('/process', methods=['POST'])
def process_database():
    """
    批量处理数据库中的新闻数据，进行分词并更新
    """
    try:
        print("[API] 开始处理数据库中的新闻数据...")
        
        # 从数据库加载数据
        segmenter.load_data(
            db_config,
            "SELECT id, content, title, description, keywords FROM feed_items"
        )
        
        if not segmenter.rows:
            return jsonify({
                "success": True,
                "message": "没有找到需要处理的数据",
                "processed": 0,
                "skipped": 0
            })
        
        print(f"[API] 从数据库加载了 {len(segmenter.rows)} 条记录")
        
        # 执行分词处理并写回数据库
        segmenter.write_back(
            db_config,
            table_name='feed_items',
            id_column='id',
            html_column='content',
            keywords_column='keywords',
            title_column='title',
            description_column='description'
        )
        
        # 从write_back方法的输出中提取处理统计信息
        # 注意：当前write_back方法只打印统计，需要修改以返回统计信息
        return jsonify({
            "success": True,
            "message": "数据库分词处理完成",
            "total_records": len(segmenter.rows)
        })
        
    except Exception as e:
        print(f"[API] 处理过程中出现错误: {e}")
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/process-single', methods=['POST'])
def process_single_html():
    """
    对单个HTML内容进行分词处理（不涉及数据库）
    """
    data = request.get_json()
    html = data.get('html', '')

    if not html:
        return jsonify({"error": "No HTML content provided"}), 400

    try:
        segmented_html = segmenter.segment_html(html)
        return jsonify({
            "success": True,
            "segmented_html": segmented_html
        })
    except Exception as e:
        return jsonify({
            "success": False,
            "error": str(e)
        }), 500

@app.route('/health', methods=['GET'])
def health_check():
    """
    健康检查接口
    """
    try:
        # 测试数据库连接
        with psycopg.connect(**db_config, client_encoding='UTF8') as conn:
            with conn.cursor() as cur:
                cur.execute("SELECT COUNT(*) FROM feed_items")
                count = cur.fetchone()[0]
        
        return jsonify({
            "status": "healthy",
            "database": "connected",
            "total_records": count
        })
    except Exception as e:
        return jsonify({
            "status": "unhealthy",
            "database": "disconnected",
            "error": str(e)
        }), 500

@app.route('/status', methods=['GET'])
def get_status():
    """
    获取数据库中的处理状态
    """
    try:
        with psycopg.connect(**db_config, client_encoding='UTF8') as conn:
            with conn.cursor() as cur:
                # 统计总记录数
                cur.execute("SELECT COUNT(*) FROM feed_items")
                total = cur.fetchone()[0]
                
                # 统计已处理的记录数（有keywords的）
                cur.execute("SELECT COUNT(*) FROM feed_items WHERE keywords IS NOT NULL AND keywords != ''")
                processed = cur.fetchone()[0]
                
                # 统计未处理的记录数
                unprocessed = total - processed
                
        return jsonify({
            "total_records": total,
            "processed_records": processed,
            "unprocessed_records": unprocessed,
            "processing_rate": f"{(processed/total*100):.1f}%" if total > 0 else "0%"
        })
    except Exception as e:
        return jsonify({
            "error": str(e)
        }), 500

if __name__ == "__main__":
    app.run(host='0.0.0.0', port=5000, debug=True)
