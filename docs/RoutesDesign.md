# 路由设计

---

### 1. 鉴权与用户中心 (Auth & Users)

| 方法 | 路径 | 描述 | 关键参数/备注 |
| --- | --- | --- | --- |
| **POST** | `/api/v1/auth/login` | 用户登录 | 返回 JWT Token |
| **POST** | `/api/v1/auth/refresh` | 刷新 Token |  |
| **GET** | `/api/v1/users/me` | 获取当前用户信息 | 包含积分、所属机构、ID |
| **PUT** | `/api/v1/users/me` | 更新个人资料 | 修改昵称、密码等 |
| **GET** | `/api/v1/users/me/points` | 查询积分余额及流水 | 关联 `points_transactions` |
| **GET** | `/api/v1/users/notifications` | 获取我的消息通知 | 系统通知、分享提醒等 |

---

### 2. 全网检索与元数据 (Search & Metadata)

此模块处理“未入库”的数据和基础字典数据。

| 方法 | 路径 | 描述 | 关键参数/备注 |
| --- | --- | --- | --- |
| **GET** | `/api/v1/search/global` | **核心：全网智能检索** | `q`:关键词, `source`:全网/库内, `agency_id`, `date_range`, `model` (basic/advanced/pro - 触发积分扣除) |

#### 参数详情设计

前端根据用户开关状态，传递不同的参数组合：

**公共参数（无论全网还是本地都需要）：**

| 参数名 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `q` | string | 是 | 检索关键词 (Keywords) |
| `source` | string | 是 | **核心开关**：`web` (全网/上网检索), `local` (本地数据库) |
| `date_start` | date | 否 | 起始日期 |
| `date_end` | date | 否 | 截止日期 |
| `agency_id` | int | 否 | 筛选特定机构 |
| `country_id` | int | 否 | 筛选特定国家 |

**模式 A：当 `source=web` (全网检索) 时的附加参数：**

> *此模式会触发积分扣除逻辑，并调用爬虫/外部API。*

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| `model` | string | **模型分级**：`basic` (不消耗配额/基础), `advanced` (高级), `professional` (专业/强力LLM) |
| `fetch_limit` | int | **数量限制**：单次抓取最大条数 (如 10, 20, 50) |
| `auto_filter` | bool | 是否开启 AI 自动过滤无关信息 (默认为 true) |

**模式 B：当 `source=local` (本地检索) 时的附加参数：**

> *此模式仅查询 `intelligences` 表，速度快，不消耗配额。*

| 参数名 | 类型 | 说明 |
| --- | --- | --- |
| `scope` | string | **检索范围**：<br>

<br>`all` (我有权限查看的所有情报，默认)<br>

<br>`mine` (仅限我收藏/入库的)<br>

<br>`team` (仅限团队库中的) |
| `has_pdf` | bool | 是否仅看有 PDF 原文的情报 |
| `sort` | string | 排序方式：`relevance` (相关度), `date_desc` (最新发布), `rating_desc` (高分优先) |

---

| **GET** | `/api/v1/search/check-duplication` | **查重检测** | `urls`: [Array] 或 `titles`: [Array]。返回库中已存在的 ID (用于前端标记绿色/黄色) |
| **GET** | `/api/v1/org/countries` | 获取国家列表 | 用于筛选下拉框 |
| **GET** | `/api/v1/org/agencies` | 获取机构列表 | `country_id`: 筛选特定国家的机构 |

---

### 3. 情报资源管理 (Intelligences)

此模块对应 `intelligences` 表，处理已入库（个人或团队）的数据。

| 方法 | 路径 | 描述 | 关键参数/备注 |
| --- | --- | --- | --- |
| **POST** | `/api/v1/intelligences` | **情报入库** | 将检索结果存入 DB。`visibility`: private (个人)/team (团队) |
| **GET** | `/api/v1/intelligences` | **情报列表查询** | `scope`: mine/team/shared, `keywords`, `has_pdf`: boolean, `sort`: date/rating |
| **GET** | `/api/v1/intelligences/{id}` | 获取情报详情 | 包含摘要、正文、标签、评分统计 |
| **DELETE** | `/api/v1/intelligences/{id}` | 删除情报 | 软删除或硬删除，需校验权限 |
| **GET** | `/api/v1/intelligences/{id}/pdf` | 下载/预览 PDF | 如果 `content` 中存储的是路径，由此接口流式返回文件 |
| **POST** | `/api/v1/intelligences/{id}/ratings` | **情报评分** | `score`: 0-5。对应 `ratings` 表 |
| **POST** | `/api/v1/intelligences/{id}/share` | **分享情报** | `target_type`: user/team, `target_id`. 写入 `intelligence_shares` 或 `permissions` |

---

### 4. 团队与协作空间 (Teams & Collaboration)

此模块对应 `teams`, `team_members` 及 ACL 逻辑。

| 方法 | 路径 | 描述 | 关键参数/备注 |
| --- | --- | --- | --- |
| **GET** | `/api/v1/teams` | 获取我的团队列表 |  |
| **POST** | `/api/v1/teams` | 创建新团队 |  |
| **GET** | `/api/v1/teams/{id}` | 获取团队详情 |  |
| **GET** | `/api/v1/teams/{id}/members` | 获取团队成员列表 |  |
| **POST** | `/api/v1/teams/{id}/members` | **添加成员** | `user_email` 或 `user_id`, `role` |
| **DELETE** | `/api/v1/teams/{id}/members/{uid}` | 移除成员 | 仅管理员可用 |
| **PUT** | `/api/v1/teams/{id}/members/{uid}` | 修改成员角色 | 修改 `role` (admin/member) |
| **GET** | `/api/v1/teams/{id}/intelligences` | **获取团队情报池** | 筛选 `permissions` 表中 subject 为该 team 的资源 |
| **POST** | `/api/v1/teams/{id}/import` | 批量导入情报到团队 | `intelligence_ids`: [Array] |

---

### 5. AI 监听与分析 (AI Monitor & Analysis)

对应功能中心的高级功能。

| 方法 | 路径 | 描述 | 关键参数/备注 |
| --- | --- | --- | --- |
| **GET** | `/api/v1/monitors` | 获取监听任务列表 |  |
| **POST** | `/api/v1/monitors` | 创建监听任务 | `keywords`, `source_agencies`, `frequency` |
| **PATCH** | `/api/v1/monitors/{id}/status` | 开启/暂停监听 |  |
| **POST** | `/api/v1/analysis/summary` | **生成综述报告** | `intelligence_ids`: [Array], `prompt_template`. 触发大模型，消耗积分 |

---

### 6. 数据流转与权限控制设计说明

为了配合你的数据库设计，后端在处理请求时需要遵循以下 **ACL (Access Control List)** 逻辑：

1. **入库逻辑 (`POST /intelligences`)**:
* 如果是“存入个人库”：
* 插入 `intelligences` 表。
* 插入 `permissions` 表：`resource_type='intelligence'`, `resource_id=new_id`, `subject_type='user'`, `subject_id=current_user_id`, `action='admin'`。


* 如果是“存入团队库”：
* 插入 `intelligences` 表。
* 插入 `permissions` 表：`subject_type='team'`, `subject_id=team_id`, `action='view'` (或其他权限)。




2. **列表查询逻辑 (`GET /intelligences`)**:
* **个人库模式**：查询 `permissions` 表中 `subject_type='user'` 且 `subject_id=me` 的 resource_id。
* **团队库模式**：查询 `permissions` 表中 `subject_type='team'` 且 `subject_id=current_team_id` 的 resource_id。
* **分享给我的**：查询 `intelligence_shares` 表中 `target_user_id=me` 的记录。


3. **积分扣除逻辑**:
* 在调用 `GET /search/global` 时，如果 `model` 参数为 `advanced` 或 `professional`，后端中间件需先检查 `users.points`，并在请求成功后写入 `points_transactions` 表（action_type=`model_call`）。



### 7. 响应体结构示例 (JSON)

**情报列表响应 (`GET /intelligences`)**

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 105,
    "page": 1,
    "items": [
      {
        "id": 1024,
        "title": "NSF 2026 财年量子计算资助规划",
        "agency_name": "National Science Foundation",
        "country_code": "US",
        "publish_date": "2025-12-15",
        "has_pdf": true,
        "my_rating": 4,  // 当前用户评分
        "avg_rating": 4.5, // 平均分
        "status_in_library": true // 前端展示去重状态
      }
    ]
  }
}

```