# 部署与 API 使用文档

本文档覆盖：后端配置、部署方法、API 接口使用、前端 HBuilderX 运行步骤。
（不含微信小程序账号注册，那部分请另见微信公众平台官方指引。）

---

## 一、环境要求

| 组件 | 版本 | 说明 |
| --- | --- | --- |
| Go | 1.22+ | 编译运行后端 |
| PostgreSQL | 14+ | 数据库（支持声明式分区） |
| HBuilderX | 最新版 | 运行/编译前端 uni-app |
| Python | 3.7+ | （可选）运行传感器模拟脚本 |

---

## 二、后端配置

### 2.1 配置文件

复制示例配置并修改：

```bash
cd backend
cp config.example.yaml config.yaml
```

`config.yaml` 字段说明：

```yaml
server:
  addr: ":8080"          # 监听地址
  mode: "debug"          # debug / release

database:
  dsn: "host=127.0.0.1 port=5432 user=postgres password=postgres dbname=deluge sslmode=disable"
  max_open_conns: 20
  max_idle_conns: 5

jwt:
  secret: "change-me-in-prod"   # 生产务必修改
  expire_hours: 168             # token 有效期(小时)

wechat:
  appid: ""              # 小程序 AppID（接入微信登录时填）
  secret: ""             # 小程序 AppSecret

readings_retention_days: 30     # 原始读数热库保留天数
async_workers: 4                # 异步任务 worker 数
device_secret: ""               # 设备上报 HMAC 密钥(空=不校验,便于联调)

tts:
  provider: "mock"       # mock / tencent
  secret_id: ""
  secret_key: ""
  region: "ap-guangzhou"

storage:
  provider: "mock"       # mock / cos
  secret_id: ""
  secret_key: ""
  region: "ap-guangzhou"
  bucket: ""
  base_url: ""           # COS 访问域名
```

### 2.2 环境变量覆盖（优先级高于配置文件）

敏感项建议用环境变量注入，避免写入文件：

| 环境变量 | 覆盖字段 |
| --- | --- |
| `SERVER_ADDR` | server.addr |
| `PG_DSN` | database.dsn |
| `JWT_SECRET` | jwt.secret |
| `WX_APPID` / `WX_SECRET` | wechat.* |
| `DEVICE_SECRET` | device_secret |
| `ASYNC_WORKERS` | async_workers |
| `READINGS_RETENTION_DAYS` | readings_retention_days |

### 2.3 数据库准备

只需创建空库，建表由程序启动时自动迁移完成：

```sql
CREATE DATABASE deluge;
```

启动后会自动执行 `migrations/` 下的 SQL（建表 + 分区 + 种子数据），无需手动建表。
种子数据包含：示范网格、洪水预警模板、4 条水位规则（蓝2.5/黄3.0/橙3.5/红4.0）、1 个示例设备。

---

## 三、后端部署

### 3.1 开发运行

```bash
cd backend
go run ./cmd/server                 # 默认读取 ./config.yaml
go run ./cmd/server -config /path/to/config.yaml   # 指定配置
```

启动成功日志：加载规则数量 → 定时任务启动 → 监听 :8080。

### 3.2 编译为单二进制

```bash
cd backend
# 当前平台
go build -o server ./cmd/server
# 交叉编译 Linux（在 Windows 上）
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o server ./cmd/server
```

部署时只需 `server` 二进制 + `config.yaml`（迁移 SQL 已内嵌进二进制）。

### 3.3 systemd 托管（Linux 生产）

`/etc/systemd/system/deluge.service`：

```ini
[Unit]
Description=Deluge Warning Backend
After=network.target postgresql.service

[Service]
WorkingDirectory=/opt/deluge
ExecStart=/opt/deluge/server -config /opt/deluge/config.yaml
Environment=PG_DSN=host=127.0.0.1 port=5432 user=deluge password=*** dbname=deluge sslmode=disable
Environment=JWT_SECRET=***
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload && sudo systemctl enable --now deluge
```

### 3.4 Docker 部署

`backend/Dockerfile`（示例，需自行创建）：

```dockerfile
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /server ./cmd/server

FROM debian:stable-slim
COPY --from=build /server /server
COPY config.yaml /config.yaml
EXPOSE 8080
ENTRYPOINT ["/server", "-config", "/config.yaml"]
```

```bash
cd backend
docker build -t deluge-backend .
docker run -d -p 8080:8080 -e PG_DSN="..." -e JWT_SECRET="..." deluge-backend
```

### 3.5 健康检查

```bash
curl http://127.0.0.1:8080/healthz
# {"status":"ok","time":"..."}
```

---

## 四、API 接口使用

基础前缀 `/api/v1`。除登录和设备上报外，其余接口需在请求头携带 `Authorization: Bearer <token>`。

### 4.1 登录（公开）

```
POST /api/v1/auth/wx-login
Body: { "code": "xxx", "role": "villager" }   # role 仅首次注册生效: villager / gridworker
返回: { "token": "...", "user": {...} }
```

> 本地联调说明：后端用 `code` 派生 openid（`openid-<code>`），无需真实微信。
> 接入微信后，把 `internal/api/handler.go` 的 `resolveOpenID` 改为调用 code2session。

### 4.2 设备读数上报（设备侧，公开 + 可选签名）

```
POST /api/v1/devices/:id/readings
Header(可选): X-Signature: HMAC_SHA256(device_secret, body) 十六进制
Body: { "device_id": "dev-water-0001", "value": 3.8, "unit": "m", "reported_at": "2025-..." }
返回: { "ok": true }
```

后端同步落库 + 规则匹配，命中则异步生成预警，立即返回（不阻塞）。

### 4.3 网格员接口（role=gridworker）

```
GET  /api/v1/devices                         设备列表
GET  /api/v1/devices/:id/readings/latest     最新读数
GET  /api/v1/devices/:id/trend?from=&to=&metric=water_level   趋势(日聚合)
GET  /api/v1/alerts?status=pending_review    预警列表(可按状态过滤)
GET  /api/v1/alerts/:id                      预警详情 + 流转日志
POST /api/v1/alerts/:id/review               复核 Body:{action:confirm|modify|cancel, content?}
POST /api/v1/alerts                          人工兜底发布 Body:{grid_id, level, content}
POST /api/v1/alerts/:id/archive              归档(仅 handled 状态)
GET  /api/v1/tasks?status=pending            我的任务
POST /api/v1/tasks/:id/confirm               确认接收
POST /api/v1/tasks/:id/handle                提交处置 Body:{remark, attachments:[{type,cos_key,url}]}
POST /api/v1/uploads/presign                 申请直传地址 Body:{key} 返回:{upload_url, access_url}
```

### 4.4 村民接口（role=villager）

```
GET  /api/v1/profile                 当前用户
GET  /api/v1/village/alerts          本网格生效预警
GET  /api/v1/alerts/:id/broadcast    广播文本 + tts_url
POST /api/v1/reports                 隐患上报 Body:{content, lng, lat}
GET  /api/v1/guides?disaster_type=flood   避险指引
```

### 4.5 完整闭环验证（curl 示例）

```bash
# 1) 注册一个网格员
curl -X POST http://127.0.0.1:8080/api/v1/auth/wx-login \
  -H "Content-Type: application/json" -d '{"code":"gw1","role":"gridworker"}'
# 记下返回的 token

# 2) 模拟设备上报触发橙色预警
curl -X POST http://127.0.0.1:8080/api/v1/devices/dev-water-0001/readings \
  -H "Content-Type: application/json" \
  -d '{"device_id":"dev-water-0001","value":3.8,"unit":"m"}'

# 3) 网格员查看预警(橙色直发,无需复核)
curl http://127.0.0.1:8080/api/v1/alerts -H "Authorization: Bearer <token>"
```

> 也可直接用 `scripts/sim_water_level.py --scenario surge --device dev-water-0001 --interval 1`
> 自动打通"上报 → 触发 → 派发"流程。

---

## 五、前端 HBuilderX 运行

前端为 uni-app 工程（`frontend/`），可编译为 H5 / 微信小程序 / App 多端。

### 5.1 准备

1. 下载安装 **HBuilderX**（App 开发版，自带 uni-app 编译器）。
2. 菜单"文件 → 打开目录"，选择本仓库的 `frontend/` 目录。
3. 修改后端地址：打开 `frontend/App.vue`，将 `globalData.baseURL` 改为后端地址
   （本机默认 `http://127.0.0.1:8080`；真机/小程序请用可公网访问的地址）。

### 5.2 运行到 H5（最快，浏览器预览，无需账号）

菜单"运行 → 运行到浏览器 → Chrome"。浏览器打开后即可用"村民/网格员"登录联调。

> H5 跨域：后端已默认放行（Gin 默认）；若浏览器报 CORS，可在后端加 CORS 中间件，
> 或用 HBuilderX 内置代理。本地同源调试建议直接用"运行到浏览器"。

### 5.3 运行到微信小程序

1. 在 `frontend/manifest.json` 的 `mp-weixin.appid` 填入你的小程序 AppID。
2. 菜单"运行 → 运行到小程序模拟器 → 微信开发者工具"。
3. HBuilderX 会编译并自动拉起微信开发者工具。
4. 微信端真机预览需在小程序后台配置服务器域名（https）。

### 5.4 运行到 App（可选）

菜单"运行 → 运行到手机或模拟器"，需连接真机或安装模拟器。
定位、录像等能力 manifest 已声明权限。

### 5.5 打包发布

- H5：菜单"发行 → 网站-H5"，产物部署到任意静态服务器。
- 小程序：菜单"发行 → 小程序-微信"，生成代码后在微信开发者工具上传审核。

---

## 六、需在云服务商获取的凭证

以下为接入生产功能所需的外部凭证（本地联调用 mock，可暂不配置）：

| 功能 | 服务商 | 获取内容 | 配置位置 |
| --- | --- | --- | --- |
| 小程序登录 | 微信公众平台 | AppID + AppSecret | `wechat.*` + 代码 resolveOpenID |
| 订阅消息推送 | 微信公众平台 | 消息模板 ID | 代码 Pusher 实现 |
| 语音合成 TTS | 腾讯云 | SecretId/Key + 开通 TTS | `tts.*`(provider=tencent) |
| 对象存储 COS | 腾讯云 | SecretId/Key + Bucket + 域名 | `storage.*`(provider=cos) |

> TTS 与 COS 共用同一对腾讯云账号密钥。替换 mock 实现的位置在
> `backend/cmd/server/main.go` 的 `buildTTS / buildStorage / buildPusher`。

---

## 七、定时任务说明（无需手动干预）

后端内置 cron（`internal/cron`），自动执行：

| 任务 | 时间 | 动作 |
| --- | --- | --- |
| 预建分区 | 每天 00:05 + 启动时 | 创建未来 3 天 readings 分区 |
| 日聚合 | 每天 00:30 | 聚合昨日读数写入 daily_water_stats |
| 清理过期分区 | 每天 01:00 | DROP 超过保留期的分区 |
| 超时检查 | 每分钟 | 检查派发超时未确认任务 |

保留天数由 `readings_retention_days` 控制（默认 30）。

---

## 八、常见问题

- **启动报连接数据库失败**：检查 `PG_DSN`/`config.yaml` 的 dsn，确认 PostgreSQL 已启动且库已创建。
- **设备上报 401**：配置了 `device_secret` 但未带正确 `X-Signature`；联调时可留空该配置。
- **小程序请求失败**：检查 `baseURL` 是否可达；微信端需配置合法域名（https）。
- **预警未自动派发**：蓝/黄级别默认需人工复核（`pending_review`），橙/红级别直发；
  可在 `rules` 表调整 `review_required`。

