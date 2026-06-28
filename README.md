# 乡村应急广播与预警信息联动小程序

面向乡村暴雨夜应急场景的预警联动系统：传感器数据自动触发预警，文本转语音(TTS)广播，
网格员处置闭环，村民及时触达。**仅设村民、网格员两个角色，网格员兼任后台管理与复核。**

## 目录结构

```
delugeWarning/
├── docs/                  PRD 与技术设计文档
│   ├── PRD.md
│   └── TECH_DESIGN.md
├── backend/               Go 后端（Gin + dbr + PostgreSQL）
│   ├── cmd/server/        程序入口
│   ├── internal/          分层代码（api/service/repository/...）
│   ├── migrations/        SQL 迁移（内嵌自动执行）
│   └── config.example.yaml
├── frontend/              uni-app 工程（HBuilderX 打开运行，村民端 + 网格员端）
└── scripts/               传感器数据模拟脚本（100 / 1000 版）
```

## 技术栈

- 后端：Go + Gin + dbr(SQL 构建器) + PostgreSQL + robfig/cron + yaml.v3
- 存储/TTS：腾讯云 COS / 腾讯云 TTS（默认 mock 实现，便于本地联调）
- 推送：微信订阅消息（默认 mock）
- 前端：uni-app（可编译为微信小程序 / H5 / App），双角色

## 快速开始

### 1. 启动 PostgreSQL 并配置

```bash
cd backend
cp config.example.yaml config.yaml
# 修改 config.yaml 中的 database.dsn，或设置环境变量 PG_DSN
```

### 2. 运行后端（自动迁移建表 + 种子数据）

```bash
cd backend
go run ./cmd/server
# 监听 :8080，启动时自动执行 migrations 并加载规则
```

### 3. 用模拟脚本触发预警闭环

```bash
cd scripts
# 暴涨场景：水位冲破红色阈值，自动触发预警
python sim_water_level.py --scenario surge --device dev-water-0001 --interval 1
```

后端会：落库读数 → 规则匹配 → 异步生成预警(含 TTS) → 直发派发任务 / 转待复核。

### 4. 运行前端（uni-app）

用 HBuilderX 打开 `frontend/` 目录，修改 `App.vue` 中 `globalData.baseURL` 指向后端地址，
然后"运行 → 运行到浏览器/微信开发者工具"。登录页可选择"村民"或"网格员"身份。
详细步骤见 `docs/DEPLOYMENT.md`。

## 核心闭环

```
传感器读数 → 规则引擎(内存,防抖) → [待复核?] → 派发任务 → 网格员确认
   → 现场处置(上传照片/视频) → 网格员归档
```

详见 `docs/TECH_DESIGN.md`。

## 数据治理（100 传感器规模）

- readings 按天分区 + BRIN 索引；热库保留 30 天
- 每日凌晨聚合为 daily_water_stats（日最高/平均/最低），过期分区 DROP
- 趋势查询走日聚合表，保证长期性能稳定

## 测试

```bash
cd backend && go test ./...   # 规则引擎、工作流状态机单元测试
```
