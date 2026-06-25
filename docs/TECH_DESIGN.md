# 技术设计文档 —— 乡村应急广播与预警信息联动小程序

## 1. 总体架构

```
          ┌────────────────────────────────────────────────┐
          │                 微信小程序前端                    │
          │   村民端视图        网格员端视图       (管理端 H5)  │
          └───────────────┬────────────────────┬───────────┘
                          │ HTTPS / JWT        │
                ┌─────────▼────────────────────▼─────────┐
                │            Go 后端（单体，分层）           │
                │  Gin 路由 → Service → Repository(dbr)    │
                │  ┌──────────────────────────────────┐   │
                │  │ 工作流引擎(状态机)  规则引擎(内存)    │   │
                │  │ 异步任务队列(TTS/推送/COS上传)        │   │
                │  │ 定时任务(分区维护/日聚合/归档/超时)    │   │
                │  └──────────────────────────────────┘   │
                └───────┬──────────┬──────────┬───────────┘
                        │          │          │
                  PostgreSQL   腾讯云 COS    微信订阅消息
                  (14+ 分区表)  (媒体+冷归档) / 腾讯云 TTS
                        ▲
                        │ HTTP上报 / MQTT（~100 设备）
                  水位/雨量传感器
```

设计原则：**单体优先、模块清晰、依赖最小**。数据库固定使用 PostgreSQL。

### 1.1 容量目标（按 100 传感器规模设计）

| 指标 | 设计取值 | 说明 |
| --- | --- | --- |
| 设备规模 | 100 个传感器 | 水位/雨量混合 |
| 采集频率 | 平峰 1 次/分钟，暴雨期可提至 1 次/10 秒 | 设备侧可配 |
| 平均写入 QPS | ≈ 1.67（100/60） | 远低于 PG 单实例上限 |
| 峰值写入 QPS | ≈ 10（暴雨提频） | 含重试仍极低 |
| 综合 QPS（含读） | 峰值几十级 | Gin 单机数千 QPS 余量充足 |
| readings 日增 | ≈ 14.4 万行/天 | 100×60×24 |
| readings 年增 | ≈ 5256 万行/年 | 仅冷归档总量，不在热库常驻 |
| 热库常驻 | 近 N 天原始读数（默认 30 天 ≈ 432 万行） | 分区裁剪 + BRIN 索引 |

> 结论：该量级在原生 PostgreSQL 舒适区内，余量极大。关键不在 QPS，而在**数据累积治理**——
> 用"分区 + 日聚合 + COS 归档"保持热库长期轻量；慢操作（TTS/推送/COS）异步化避免拖慢上报。
> TimescaleDB 列为升级路径：仅当扩展到 ≥5000 设备或秒级采集（数十亿行/年）时再启用。

## 2. 技术选型

| 层 | 选型 | 说明 |
| --- | --- | --- |
| 语言 | Go 1.22+ | 单二进制、低内存、易部署 |
| Web 框架 | Gin | 轻量、生态成熟 |
| 数据访问 | **dbr**（gocraft/dbr）+ pgx/lib/pq 驱动 | 轻量 SQL 构建器，无 ORM 反射开销，贴近 SQL |
| 数据库 | **PostgreSQL 14+** | 生产与开发统一；readings 用声明式分区 + BRIN 索引 |
| 数据库迁移 | golang-migrate（纯 SQL 脚本） | 版本化建表，dbr 不做自动迁移 |
| 分区管理 | **pg_partman**（推荐）或自写 cron | 自动建未来分区 + DROP 过期分区 |
| 缓存/调度 | 进程内（go-cache + robfig/cron） | 避免引入 Redis，保持轻量 |
| 异步任务 | 进程内 worker pool + channel 队列 | TTS/推送/COS 上传异步化，DB 持久化补偿 |
| 鉴权 | JWT（golang-jwt） | 无状态、角色声明 |
| 配置 | gopkg.in/yaml.v3 + env 覆盖 | 轻量 YAML 解析，无 viper 重依赖 |
| 对象存储 | **腾讯云 COS**（cos-go-sdk-v5） | 存储照片/视频/TTS音频；接口抽象便于切换 |
| TTS | 接口抽象：云 TTS REST（腾讯云/讯飞/百度） | 优先腾讯云 TTS，生成 mp3 存 COS |
| 推送 | 微信订阅消息 API | 模板消息推送 |
| 设备接入 | HTTP（默认）/ MQTT（可选） | MQTT 用 mochi-mqtt 内嵌 |
| 前端 | 微信原生小程序 / UniApp | 双角色单包，按 role 分包 |
| 部署 | Docker / 单二进制 + systemd | 一键部署 |

> 轻量化关键：用 **dbr** SQL 构建器替代 GORM，去掉重反射与魔法；不强依赖 Redis/Kafka；
> 调度、防抖、缓存均进程内实现；存储/TTS 走腾讯云，本地仅留抽象接口便于联调。

## 3. 后端模块划分

```
/cmd/server          程序入口
/internal
  /api               Gin handler（路由、参数校验、DTO）
  /service           业务逻辑
    alert            预警发布/查询
    workflow         工作流状态机（核心）
    rule             规则引擎（内存规则匹配/防抖/升级）
    device           设备与读数接入（落库+入队）
    tts              文本转语音抽象
    push             微信订阅消息推送
    task             网格员任务处置
    report           村民隐患上报
    stats            日聚合查询（趋势/统计，读 daily_water_stats）
  /async             进程内任务队列 + worker pool（TTS/推送/COS 上传）
  /cron              定时任务（分区维护/日聚合/冷归档/超时重派/推送补偿）
  /repository        dbr 数据访问（SQL 构建 + 映射）
  /model             实体定义（struct + db tag）
  /middleware        JWT、日志、限流、恢复
  /pkg               wx登录、cos storage、tts client、id生成
/config              配置文件
/migrations          golang-migrate SQL 迁移脚本（up/down，含分区父表与聚合表）
```

## 4. 数据模型（核心表）

```sql
-- 用户（村民/网格员；网格员兼后台管理与复核）role: villager / gridworker
users(id, openid, role, name, phone, grid_id, created_at)

-- 网格/区域
grids(id, name, village, manager_user_id, geojson)

-- 设备
devices(id, type, name, grid_id, status, last_value, last_report_at)

-- 设备读数（时序大表，按天范围分区，详见 §4.1）
readings(id, device_id, value, unit, reported_at)   -- PARTITION BY RANGE(reported_at)

-- 日聚合趋势表（降采样，永久保留，趋势图/统计数据源，详见 §4.2）
daily_water_stats(device_id, metric, stat_date, max_value, avg_value,
                  min_value, sample_cnt, unit)      -- PK(device_id, metric, stat_date)

-- 预警规则（review_required: 命中后是否需人工复核再发布）
rules(id, device_type, metric, operator, threshold, level, cooldown_sec,
      template_id, review_required, enabled)

-- 预警模板
templates(id, name, disaster_type, content_tpl, enabled)

-- 预警（工作流主单）source: sensor(系统自动) / manual(人工兜底)
alerts(id, source, level, disaster_type, grid_id, title, content,
       tts_url, status, triggered_by, reviewed_by, created_at, archived_at)

-- 工作流流转日志
alert_logs(id, alert_id, from_status, to_status, operator_id, remark, created_at)

-- 网格员处置任务
tasks(id, alert_id, assignee_id, status, handle_remark, finished_at)

-- 处置证据（照片/视频）cos_key: COS 对象键；url 可由 key + 域名拼接或临时签发
attachments(id, task_id, type, cos_key, url, created_at)

-- 村民隐患上报
reports(id, reporter_id, grid_id, content, lng, lat, status, created_at)
report_attachments(id, report_id, type, cos_key, url)
```

`alerts.status` 枚举：`pending_review/triggered/dispatched/confirmed/handled/archived/canceled`。

### 4.1 readings 时序表分区（按天）

100 设备分钟级采集 → 14.4 万行/天、5256 万行/年。采用**声明式范围分区（按天）**：

```sql
CREATE TABLE readings (
    id          BIGINT GENERATED ALWAYS AS IDENTITY,
    device_id   TEXT        NOT NULL,
    value       NUMERIC(8,3) NOT NULL,
    unit        TEXT,
    reported_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (reported_at);

-- 每天一个子分区（由 pg_partman 或 cron 自动创建未来分区）
-- 例：CREATE TABLE readings_20250601 PARTITION OF readings
--       FOR VALUES FROM ('2025-06-01') TO ('2025-06-02');

-- 时序数据用 BRIN 索引（按 reported_at 物理有序，比 B-tree 省 90%+ 空间）
CREATE INDEX ON readings USING BRIN (reported_at);
-- 单设备近期查询用普通索引（按需，仅建在近月分区或父表）
CREATE INDEX ON readings (device_id, reported_at DESC);
```

- 分区粒度：**按天**（与大规模一致，便于未来平滑扩展）。单分区约 14.4 万行，
  查询裁剪精准，DROP 旧分区瞬时完成、无 vacuum 膨胀。
  （注：100 设备量级亦可改按月分区，单分区约 432 万行，进一步减少分区数量。）
- 索引：BRIN 为主（时序追加写入、物理有序，极省空间）；按需补 `(device_id, reported_at)`。
- 写入：父表插入由 PG 自动路由到当天分区。

### 4.2 日聚合表与冷数据治理

热库只保留近期原始读数，过期前**先降采样到日聚合表**，再 DROP 原始分区。

```sql
CREATE TABLE daily_water_stats (
    device_id   TEXT         NOT NULL,
    metric      TEXT         NOT NULL,          -- water_level / rainfall
    stat_date   DATE         NOT NULL,          -- 本地时区(Asia/Shanghai)的日期
    max_value   NUMERIC(8,3) NOT NULL,          -- 当日最高（防汛关键指标）
    avg_value   NUMERIC(8,3) NOT NULL,          -- 当日平均
    min_value   NUMERIC(8,3),                   -- 当日最低
    sample_cnt  INTEGER      NOT NULL,          -- 当日样本数（数据质量参考）
    unit        TEXT,
    PRIMARY KEY (device_id, metric, stat_date)  -- 天然幂等，重复聚合不出错
);
```

冷热分层策略：

| 数据 | 热层（PG 在线） | 冷层（腾讯云 COS） | 永久保留 |
| --- | --- | --- | --- |
| readings 原始 | 近 30 天（约 432 万行） | 超期分区导出 Parquet 归档 | 否（COS 生命周期转低频/归档存储） |
| daily_water_stats | 全量常驻（≈3.65 万行/年，极小） | — | 是 |
| alerts/日志/证据 | 全量在线 | 媒体存 COS | 是（应急追溯） |

定时治理动作（详见 §6.1，由 cron 每日执行）：先聚合 → 导出 COS → DROP 过期分区。

## 5. 工作流引擎设计（轻量状态机）

采用**配置化状态机**，而非引入重量级 BPMN：

```go
// 合法流转表
var transitions = map[Status][]Status{
    PendingReview: {Triggered, Canceled},   // 待人工复核（低级别可选）
    Triggered:     {Dispatched, Canceled},
    Dispatched:    {Confirmed, Dispatched}, // 重派
    Confirmed:     {Handled},
    Handled:       {Archived},
}
// Transition 校验合法性 → 持久化 alert.status → 写 alert_logs → 触发副作用(推送)
```

- 每次流转：校验 → 更新状态 → 写流转日志 → 执行 Hook（推送/通知）。
- 超时控制：cron 每分钟扫描 `dispatched` 超时未确认的单，自动重派或升级提醒。
- 幂等：流转接口带 `expected_from_status`，避免并发重复操作。

## 6. 规则引擎与自动触发（系统为主发布途径）

预警**主要由系统自动产生**：传感器读数经规则引擎匹配后直接生成并发布预警，
人工仅作校验/复核与兜底补发。

```
POST /readings（同步路径，必须快，目标 < 10ms）
  1. 校验签名 + 落库（插入当天分区，单条 insert）
  2. ruleEngine.Evaluate(reading)   // 规则常驻内存，纯比较，不查库
       for rule in 内存规则[device_type]:
          if compare(value, op, threshold):
             if 冷却期内已触发同级(内存记录) -> skip(防抖)
             else -> 入队 alertJob{rule, reading}   // 仅入队，立即返回
  3. 立即响应设备 200

异步 worker（消费 alertJob，慢操作都在这里）：
   渲染模板 -> 调用腾讯云 TTS 合成 -> 上传 COS
   if rule.review_required: 创建 alert(pending_review)   // 低级别可选，等人工复核
   else:                    创建 alert(triggered) -> workflow.Dispatch(alert)
   Dispatch 内再异步执行微信推送（失败入重试队列）
```

写入路径设计要点（100 设备 / 峰值 ~10 QPS，且为未来扩展预留）：
- **同步只做"落库 + 内存规则比较 + 入队"**，亚毫秒级；TTS/推送/COS 上传全部异步。
- **规则常驻内存**：启动加载全部启用规则，规则变更时刷新，评估不查库。
- 防抖：基于 `device_id + level` 的**进程内**冷却记录，`cooldown_sec` 内忽略，省去无谓计算与重复 TTS。
- 升级：高级别规则命中时若已有低级别在途预警，则升级并重新派发。
- 发布模式：规则级 `review_required` 控制——默认**直发**，低级别可开启**人工复核**。
- 人工兜底：网格员手动创建的预警同样进入该状态机（`triggered` 或 `pending_review`）。
- 异步队列：进程内 worker pool + channel；任务先持久化（或失败重试表）保证不丢，支持补偿重试。

### 6.1 定时任务（robfig/cron 进程内调度）

| 任务 | 频率 | 动作 |
| --- | --- | --- |
| 预分区 | 每天 | 创建未来 N 天的 readings 分区（pg_partman 或自写） |
| 日聚合 | 每天 00:30 | 聚合"昨天及更早未聚合"的读数写入 daily_water_stats（幂等 UPSERT） |
| 冷归档 | 每天 | 将超 30 天的分区导出 Parquet 上传 COS |
| 清理原始 | 每天 | 确认已聚合 + 已归档后，DROP 过期 readings 分区 |
| 超时重派 | 每分钟 | 扫描 `dispatched` 超时未确认的预警，自动重派/升级提醒 |
| 推送补偿 | 每分钟 | 重试失败的微信推送 |

日聚合 + 清理的核心 SQL（按本地时区分日、先聚合后删，幂等安全）：

```sql
-- ① 聚合（只处理过去日期，绝不动当天；重复跑安全）
INSERT INTO daily_water_stats
    (device_id, metric, stat_date, max_value, avg_value, min_value, sample_cnt, unit)
SELECT r.device_id, d.metric,
       (r.reported_at AT TIME ZONE 'Asia/Shanghai')::date AS stat_date,
       MAX(r.value), AVG(r.value), MIN(r.value), COUNT(*), MAX(r.unit)
FROM readings r JOIN devices d ON d.id = r.device_id
WHERE (r.reported_at AT TIME ZONE 'Asia/Shanghai')::date
        < (now() AT TIME ZONE 'Asia/Shanghai')::date
GROUP BY r.device_id, d.metric, stat_date
ON CONFLICT (device_id, metric, stat_date) DO UPDATE
SET max_value = EXCLUDED.max_value, avg_value = EXCLUDED.avg_value,
    min_value = EXCLUDED.min_value, sample_cnt = EXCLUDED.sample_cnt;

-- ② 清理：确认已聚合且已归档后，按分区 DROP（瞬时、无膨胀）
-- DROP TABLE IF EXISTS readings_20250501;   -- 由 cron 计算过期分区名
```

## 7. 关键接口设计（REST，前缀 /api/v1）

### 通用
- `POST /auth/wx-login` 小程序 code 换 openid + JWT
- `GET  /profile` 当前用户信息

### 设备/读数（设备或网格员）
- `POST /devices/{id}/readings` 设备上报读数（签名校验）→ 同步落库+入队，立即返回
- `GET  /devices` 设备列表与状态
- `GET  /devices/{id}/trend?from=&to=` 设备水位趋势（**读 daily_water_stats 日聚合表**，非原始读数）
- `GET  /devices/{id}/readings/latest` 设备最新实时读数（读热分区）

### 预警（网格员 —— 以校验复核为主）
- `GET  /alerts` 预警列表（按状态/级别/区域，含 `pending_review`）
- `GET  /alerts/{id}` 预警详情 + 流转日志
- `POST /alerts/{id}/review` **复核系统预警**：confirm（发布）/ modify（改文案）/ cancel（撤销误报）
- `POST /alerts` 人工兜底发布预警（含生成 TTS，仅兜底场景）
- `POST /alerts/{id}/archive` 审核归档

### 网格员任务
- `GET  /tasks?status=` 我的任务
- `POST /tasks/{id}/confirm` 确认接收
- `POST /tasks/{id}/handle` 提交处置（remark + 附件 key 列表）
- `POST /uploads/presign` 申请 COS 预签名直传地址（前端直传，返回最终访问 url）

### 村民
- `GET  /village/alerts` 本区域生效预警
- `GET  /alerts/{id}/broadcast` 获取广播文本与 tts_url
- `POST /reports` 提交隐患上报
- `GET  /guides?disaster_type=` 避险指引

## 8. TTS、存储与推送抽象

```go
// 文本转语音：默认腾讯云 TTS，合成后上传 COS 返回可访问 URL
type TTSEngine interface {
    Synthesize(ctx context.Context, text string, opt Options) (audioURL string, err error)
}
// 对象存储：默认腾讯云 COS（亦可换 S3/OSS/本地）
type Storage interface {
    Put(ctx context.Context, key string, r io.Reader, contentType string) (url string, err error)
    PresignPut(ctx context.Context, key string, ttl time.Duration) (uploadURL string, err error)
}
// 推送：微信订阅消息
type Pusher interface {
    Push(ctx context.Context, openid string, tmpl TemplateMsg) error
}
```
- 存储：默认 **腾讯云 COS**（`cos-go-sdk-v5`），照片/视频/TTS音频统一存 COS；
  前端大文件可用 `PresignPut` 直传 COS，减轻后端带宽。
- TTS：优先 **腾讯云 TTS** REST，合成 mp3 → 上传 COS → 返回 URL。
- 推送：微信订阅消息；推送失败入重试队列（进程内 + DB 持久化补推）。

## 9. 前端小程序设计（双角色）

### 9.1 角色路由
- 登录后由后端返回 `role`，前端按角色加载不同 TabBar：
  - **村民**：首页(预警)、广播、避险指引、上报、我的
  - **网格员**：任务、地图、上报核查、我的
- 使用**分包加载**减少主包体积：`villager` 分包 / `gridworker` 分包。

### 9.2 关键页面
| 角色 | 页面 | 要点 |
| --- | --- | --- |
| 村民 | 预警首页 | 当前级别大字 + 颜色，倒计时/状态 |
| 村民 | 广播详情 | TTS 音频播放器（innerAudioContext） |
| 村民 | 隐患上报 | 拍照/录像 + getLocation 定位 |
| 网格员 | 任务列表 | 状态分组、红点提醒 |
| 网格员 | 任务处置 | 确认按钮、上传多媒体、提交 |

### 9.3 推送订阅
- 进入小程序时引导 `requestSubscribeMessage` 授权预警/任务模板。
- 网格员任务派发、村民预警发布触发订阅消息推送。

## 11. 性能与容量保障（100 传感器）

| 关注点 | 风险 | 应对 |
| --- | --- | --- |
| 写入吞吐 | 峰值 ~10 QPS | PG 单实例余量极大；同步路径仅落库+入队 |
| 单请求耗时 | TTS/推送拖慢上报 | 慢操作全异步，同步路径目标 < 10ms |
| 数据膨胀 | 年增 5256 万行 | 按天分区 + 30 天保留 + DROP 旧分区 |
| 查询变慢 | 大表扫描 | 趋势读日聚合表；明细查询走分区裁剪 + BRIN |
| vacuum 压力 | DELETE 产生死元组 | 用 DROP 分区替代 DELETE，零膨胀 |
| 推送/TTS 抖动 | 第三方不稳定 | 异步队列 + 失败持久化 + 定时补偿重试 |

压测验证：用 `scripts/sim_water_level.py --devices 100` 并发压测，观察实时 QPS、
成功率与 P50/P95/P99 延迟（脚本已支持，详见 `scripts/README.md`）。
更大规模（1000 设备）的压测可用 `scripts/sim_water_level_1000.py`。

升级路径：当扩展到 ≥5000 设备或秒级采集（数十亿行/年）时，将 readings 迁移到
**TimescaleDB hypertable**，启用原生列式压缩、连续聚合（替代日聚合 cron）与保留策略；
应用层接口与聚合表对外契约不变，平滑切换。

## 12. 安全与鉴权
- JWT 中携带 `user_id + role`，中间件校验角色访问权限。
- 设备上报使用设备密钥 HMAC 签名，防伪造。
- 上传：后端签发 COS 预签名地址，限定 object key 前缀、文件类型与大小、有效期；
  COS Bucket 设为私有读，访问通过临时签名 URL，避免证据被公开枚举。

## 13. 部署与运行
- 开发：`go run ./cmd/server`，连接本地/云 PostgreSQL；首次运行 `migrate up` 建表。
- 生产：`docker build` → 单容器 + 托管 PostgreSQL；对象存储用腾讯云 COS。
- 数据库迁移：golang-migrate 管理 `/migrations`（含 readings 分区父表、daily_water_stats）。
- 分区初始化：部署时预建近期+未来若干天分区；cron 持续滚动创建/清理。
- 配置项：PG_DSN、READINGS_RETENTION_DAYS（默认 30）、JWT_SECRET、
  WX_APPID/SECRET、TTS_*（腾讯云密钥）、COS_*（SecretId/SecretKey/Region/Bucket）、
  ASYNC_WORKERS（异步 worker 数）。
- 监控：结构化日志 + /healthz 健康检查 + 异步队列积压/分区数量指标。

## 14. 目录与里程碑落地
1. 搭骨架：Gin + dbr + PostgreSQL + golang-migrate + JWT + wx-login。
2. readings 分区父表 + daily_water_stats + 设备读数接入（落库+入队）。
3. 规则引擎（内存）+ 异步队列 + 工作流状态机（自动触发闭环）。
4. cron 定时任务：分区维护 + 日聚合 + 冷归档 + 超时重派。
5. 腾讯云 TTS + COS 存储 + 微信推送接入。
6. 网格员任务处置 + COS 预签名直传。
7. 村民端预警/广播/趋势/上报；管理端复核与监控、归档。
