import psycopg
from bs4 import BeautifulSoup, NavigableString
import jieba
import jieba.analyse
import html
import re

class HtmlTextSegmenter:
    def __init__(self):
        self.rows = []

    def load_data(self, db_config, sql):
        with psycopg.connect(**db_config, client_encoding='UTF8') as conn:
            with conn.cursor() as cur:
                cur.execute(sql)
                self.rows = cur.fetchall()
                for i, row in enumerate(self.rows):
                    print(f"Row {i} id={row[0]} type(content)={type(row[1])} type(title)={type(row[2])} type(description)={type(row[3])} repr content={repr(row[1])[:100]}")

    def segment_text(self, text):
        if not text.strip():
            return text
        tokens = jieba.cut(text, cut_all=False)
        result = ""
        for token in tokens:
            if re.match(r'^[\u4e00-\u9fa5A-Za-z0-9]+$', token):  # 是词语
                result += '^' + token
            else:  # 是标点或特殊字符
                result += '*' + token
        return result

    def segment_html(self, html_content, token_id=None):
        raw_html = html.unescape(html_content or "")
        soup = BeautifulSoup(raw_html, "html.parser")

        # 完全跳过这些标签（内容一起丢弃）
        skip_tags = {'script', 'style', 'iframe'}

        # 保留这些标签
        keep_tags = {'p', 'br', 'img', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6'}

        def recurse(node):
            for child in list(node.children):
                if isinstance(child, NavigableString):
                    text = str(child)
                    if text.strip():
                        segmented = self.segment_text(text)
                        child.replace_with(BeautifulSoup(segmented, "html.parser"))
                elif child.name in skip_tags:
                    child.decompose()
                elif child.name in keep_tags:
                    if child.name != 'img':
                        recurse(child)
                    # img 是自闭合标签，不处理 children
                else:
                    # unwrap 标签，但递归处理内容
                    recurse(child)
                    child.unwrap()

        recurse(soup)
        return str(soup), token_id

    def remove_span_tags(self, html_content):
        soup = BeautifulSoup(html.unescape(html_content or ""), "html.parser")
        for span in soup.find_all('span', attrs={"data-id": True}):
            span.unwrap()
        return soup.get_text(separator=" ", strip=True)

    def extract_keywords(self, html_content, title="", description=""):
        clean_html_text = self.remove_span_tags(html_content)
        clean_title_text = self.remove_span_tags(title)
        clean_description_text = self.remove_span_tags(description)

        full_text = ' '.join(filter(None, [clean_title_text, clean_description_text, clean_html_text]))
        print("[extract_keywords] 合并纯文本:", repr(full_text[:200]))

        keywords = jieba.analyse.extract_tags(full_text, topK=5, withWeight=True)
        print("[extract_keywords] 提取关键词:", keywords)
        return keywords

    def write_back(self, db_config, table_name, id_column, html_column, keywords_column, title_column, description_column):
        with psycopg.connect(**db_config, client_encoding='UTF8') as conn:
            with conn.cursor() as cur:
                for row in self.rows:
                    try:
                        record_id = row[0]
                        raw_html = row[1] or ""
                        raw_title = row[2] or ""
                        raw_description = row[3] or ""
                        _ = row[4]  # 原有关键词

                        print(f"Processing record id={record_id}")

                        def has_been_segmented(content):
                            soup = BeautifulSoup(html.unescape(content or ""), "html.parser")
                            return bool(soup.find("span", attrs={"data-id": True}))

                        token_id = 0
                        new_title, token_id = self.segment_html(raw_title, token_id) if not has_been_segmented(raw_title) else (raw_title, token_id)
                        new_description, token_id = self.segment_html(raw_description, token_id) if not has_been_segmented(raw_description) else (raw_description, token_id)
                        new_html, token_id = self.segment_html(raw_html, token_id) if not has_been_segmented(raw_html) else (raw_html, token_id)

                        keywords_with_weight = self.extract_keywords(new_html, new_title, new_description)
                        keywords_list = [kw for kw, _ in keywords_with_weight]
                        keywords_str = keywords_list  # 直接存 list（数组类型）

                        sql = f"""
                            UPDATE {table_name} SET
                                {html_column} = %s,
                                {keywords_column} = %s,
                                {title_column} = %s,
                                {description_column} = %s
                            WHERE {id_column} = %s
                        """
                        cur.execute(sql, (new_html, keywords_str, new_title, new_description, record_id))
                    except Exception as e:
                        print(f"Error processing id={record_id}: {e}")
                conn.commit()