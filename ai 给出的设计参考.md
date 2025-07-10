好的，我们来基于这两份文档，设计一个既能满足复杂实验需求，又能有效控制存储和服务器压力的数据库方案。

核心挑战分析：
1.  **数据类型的二元性**：项目数据分为两类。一类是关系型数据，如用户信息、新闻内容、实验分组等，数据量不大但关系复杂。另一类是海量时序数据，即用户的眼动、鼠标、滚动行为，其特点是写入频繁、数据量巨大（20Hz * 3600s * 10万用户 = 72亿条/天，这是个恐怖的数字）。
2.  **性能压力**：QPS 100+ 的要求本身不低，但真正的瓶颈在于每秒 `20 * 10万 = 200万` 次的潜在数据点生成，如果直接写入数据库，任何传统关系型数据库都无法承受。
3.  **A/B Test 的灵活性**：需要轻松地将用户分配到不同实验组，并能在数据分析时清晰地将行为数据与实验组关联起来。

基于以上分析，单一的数据库解决方案（如仅使用 MySQL 或 PostgreSQL）是不可行的。最佳实践是采用**混合数据库架构（Polyglot Persistence）**。

---

### 整体架构与技术选型

1.  **主数据库 (关系型)**：使用 **PostgreSQL**。
    *   **原因**：功能强大，事务支持完善，非常适合管理用户、新闻、邀请码、实验分组等结构化数据。其对 JSONB 的原生支持也为存储一些半结构化数据提供了便利。
2.  **行为日志数据库 (时序/分析型)**：使用 **ClickHouse** 或 **InfluxDB**。
    *   **原因**：这类数据库专为海量时序数据的高速写入和快速分析而设计。
        *   **ClickHouse**：列式存储，数据压缩比极高（能有效解决存储压力），查询分析性能极强，非常适合事后对海量行为数据进行聚合分析。
        *   **InfluxDB**：专业的时序数据库，对时间线数据写入和查询做了极致优化。
    *   **本方案推荐使用 ClickHouse**，因为它在数据压缩和复杂分析查询方面更具优势，更贴合本项目的研究性质。

3.  **数据采集策略 (关键)**：
    *   **前端批量上报**：前端JS绝不能每20ms发送一次请求。应在本地缓存数据，例如**每5-10秒**将积累的几百个数据点打包成一个JSON数组，通过一次HTTP请求发送到后端。这能将写入请求的频率降低数百倍，是**减轻服务器压力的核心**。
    *   **后端异步写入**：后端API接收到批量数据后，不应立即阻塞并写入ClickHouse。应先将数据快速推送到一个消息队列（如 **Redis Stream** 或 **Kafka**），然后由一个独立的消费者服务（Worker）从队列中批量拉取数据，再写入ClickHouse。这实现了流量削峰，保证了API接口的快速响应。

---

### 详细数据库表设计

#### Part 1: PostgreSQL 表设计 (主业务库)

**1. `users` - 用户信息表**
*   存储用户的基本信息和实验分组情况。

| 字段名                  | 数据类型     | 约束/索引                                   | 备注                                                         |
| :---------------------- | :----------- | :------------------------------------------ | :----------------------------------------------------------- |
| `id`                    | SERIAL       | PRIMARY KEY                                 | 自增用户ID                                                   |
| `uuid`                  | UUID         | UNIQUE, NOT NULL, DEFAULT gen_random_uuid() | 用户唯一标识符，不暴露自增ID                                 |
| `invitation_code_id`    | INT          | FOREIGN KEY -> invitation_codes(id)         | 关联使用的邀请码                                             |
| `experiment_group_json` | JSONB        |                                             | 存储用户所属的实验组，`{"recommendation_test": "A", "social_info_test": "B"}`，灵活支持多项A/B测试 |
| `birth_date`            | DATE         |                                             | 出生年月                                                     |
| `gender`                | VARCHAR(10)  |                                             | 性别                                                         |
| `education_level`       | VARCHAR(50)  |                                             | 学历                                                         |
| `residence`             | VARCHAR(100) |                                             | 居住地                                                       |
| `weekly_reading_hours`  | INT          |                                             | 每周新闻阅读时长                                             |
| `news_platforms`        | JSONB        |                                             | 经常使用的新闻平台，如 `["微信新闻", "今日头条"]`            |
| `is_active_searcher`    | BOOLEAN      |                                             | 是否主动搜索新闻                                             |
| `vision_status`         | VARCHAR(50)  |                                             | 视力状况 ('正常', '近视', '远视')                            |
| `is_vision_corrected`   | BOOLEAN      |                                             | 矫正视力是否正常                                             |
| `is_colorblind`         | BOOLEAN      |                                             | 是否色盲                                                     |
| `created_at`            | TIMESTAMPTZ  | DEFAULT NOW()                               | 账户创建时间                                                 |

**2. `invitation_codes` - 邀请码表**
*   管理邀请码的发放和使用状态。

| 字段名            | 数据类型     | 约束/索引                          | 备注                              |
| :---------------- | :----------- | :--------------------------------- | :-------------------------------- |
| `id`              | SERIAL       | PRIMARY KEY                        | 自增ID                            |
| `code`            | VARCHAR(50)  | UNIQUE, NOT NULL                   | 邀请码字符串，如 `1234-5678-ABCD` |
| `is_used`         | BOOLEAN      | NOT NULL, DEFAULT FALSE            | 是否已被使用                      |
| `used_by_user_id` | INT          | FOREIGN KEY -> users(id), NULLABLE | 使用该码的用户ID                  |
| `used_at`         | TIMESTAMPTZ  |                                    | 使用时间                          |
| `email_to`        | VARCHAR(255) |                                    | 发送到的目标邮箱（用于记录）      |
| `created_at`      | TIMESTAMPTZ  | DEFAULT NOW()                      | 创建时间                          |
| `expires_at`      | TIMESTAMPTZ  |                                    | 过期时间                          |

**3. `news_articles` - 新闻文章表**
*   存储从RSS源或其他API抓取的新闻内容。基于文档中的 `feeds` 和 `feed_items` 表合并优化。

| 字段名            | 数据类型     | 约束/索引        | 备注                           |
| :---------------- | :----------- | :--------------- | :----------------------------- |
| `id`              | SERIAL       | PRIMARY KEY      | 自增ID                         |
| `guid`            | VARCHAR(512) | UNIQUE, NOT NULL | 文章全局唯一ID，防止重复抓取   |
| `source_url`      | VARCHAR(512) | NOT NULL         | 文章原始链接                   |
| `source_name`     | VARCHAR(255) |                  | 来源名称, e.g., "XX新闻"       |
| `title`           | VARCHAR(255) | NOT NULL         | 标题                           |
| `summary`         | TEXT         |                  | 摘要                           |
| `content_text`    | TEXT         |                  | **纯文本内容**，用于分词和分析 |
| `content_html`    | TEXT         |                  | 原始HTML内容，用于展示         |
| `cover_image_url` | VARCHAR(512) |                  | 封面图链接                     |
| `published_at`    | TIMESTAMPTZ  |                  | 发布时间                       |
| `created_at`      | TIMESTAMPTZ  | DEFAULT NOW()    | 抓取入库时间                   |

**4. `user_discreet_actions` - 用户离散行为表**
*   记录用户的低频、关键行为，如点击、点赞等。

| 字段名        | 数据类型    | 约束/索引                      | 备注                                                         |
| :------------ | :---------- | :----------------------------- | :----------------------------------------------------------- |
| `id`          | BIGSERIAL   | PRIMARY KEY                    |                                                              |
| `user_id`     | INT         | FK -> users(id), INDEX         | 用户ID                                                       |
| `article_id`  | INT         | FK -> news_articles(id), INDEX | 文章ID                                                       |
| `action_type` | VARCHAR(50) | NOT NULL, INDEX                | 行为类型, 'view', 'like', 'share', 'comment', 'change_batch' |
| `session_id`  | UUID        | INDEX                          | 唯一会话ID，用于串联一次访问的所有行为                       |
| `created_at`  | TIMESTAMPTZ | DEFAULT NOW()                  | 行为发生时间                                                 |

---

#### Part 2: ClickHouse 表设计 (行为日志库)

**`gaze_events` - 眼动及连续行为事件表**
*   这是承载海量数据的核心表。使用ClickHouse的 `MergeTree` 引擎。

```sql
CREATE TABLE gaze_events (
    -- 时间戳是时序数据的核心
    timestamp         DateTime64(3, 'UTC'), -- 毫秒级精度

    -- 维度/标签 (用于过滤和分组，应为低基数)
    session_id        UUID,           -- 会话ID，关联一次完整的阅读过程
    user_id           UUID,           -- 用户UUID，关联到PostgreSQL中的用户
    article_id        Int32,          -- 新闻文章ID
    
    -- 实验分组信息 (冗余存储，便于分析)
    recommend_group   String,         -- e.g., 'random', 'collaborative_filter'
    social_info_group String,         -- e.g., 'show', 'hide'

    -- 事件类型
    event_type        Enum8('gaze' = 1, 'scroll' = 2, 'mouse' = 3), -- 事件类型

    -- 事件具体数据 (高基数列，不适合做索引)
    gaze_x            Float32,        -- 眼动坐标X (event_type='gaze'时有效)
    gaze_y            Float32,        -- 眼动坐标Y (event_type='gaze'时有效)
    
    scroll_top        UInt32,         -- 滚动位置 (event_type='scroll'时有效)
    
    mouse_x           UInt32,         -- 鼠标坐标X (event_type='mouse'时有效)
    mouse_y           UInt32,         -- 鼠标坐标Y (event_type='mouse'时有效)

    -- 客户端信息
    screen_width      UInt16,
    screen_height     UInt16
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp) -- 按月分区，便于数据管理
ORDER BY (user_id, session_id, timestamp); -- 排序键，优化查询性能
```

**为什么这样设计能解决压力问题：**
1.  **存储压力**：ClickHouse的列式存储和高压缩率可以将TB级的原始数据压缩到GB级，大大降低存储成本。按月分区也便于未来按时间归档或删除旧数据。
2.  **服务器压力（写入）**：前端批量上报 + 后端消息队列异步写入，将海量瞬时请求削平为平稳的、批量的数据库写入操作，保护了API服务器和数据库。
3.  **服务器压力（查询）**：
    *   简单的用户、新闻查询在PostgreSQL上进行，快速高效。
    *   复杂的海量行为分析查询在ClickHouse上进行。其`MergeTree`引擎和排序键设计，使得对特定用户、特定会话在某时间段内的行为查询非常快。分析时无需JOIN PostgreSQL，因为关键的实验分组信息已经冗余存入，实现了**读时解耦**。

### 总结

这个**PostgreSQL + ClickHouse + 消息队列**的混合架构方案，通过“专业分工”的方式，让每种技术都做自己最擅长的事：
*   **PostgreSQL** 负责维护业务数据的一致性和完整性。
*   **ClickHouse** 负责吞吐和分析海量的行为日志。
*   **前端批量上报**和**消息队列**作为缓冲层，保护整个系统免受高并发写入的冲击。

这样，不仅能完全满足项目的功能和A/B测试需求，还能从根本上解决海量眼动数据带来的存储和服务器压力问题，为后续的研究分析提供了高性能、高可扩展性的数据基础。