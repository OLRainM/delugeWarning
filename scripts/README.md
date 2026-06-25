# 传感器数据模拟脚本说明

模拟乡村暴雨夜场景下水位/雨量传感器向后端定时上报读数，驱动后端规则引擎**自动触发预警**，
便于联调与演示。纯 Python 标准库实现，无需安装依赖。

两个脚本**处理逻辑完全一致**，仅默认参数不同：

| 脚本 | 面向规模 | 默认模式 | 说明 |
| --- | --- | --- | --- |
| `sim_water_level.py` | **100 传感器**（系统默认部署规模） | 单设备演示 | 日常联调、功能演示 |
| `sim_water_level_1000.py` | **1000 传感器**（扩展压测） | 1000 设备并发压测 | 大规模吞吐压测 |

## 运行环境
- Python 3.7+
- 后端已启动并提供 `POST /api/v1/devices/{id}/readings` 接口

## 快速开始

`sim_water_level.py` 有两种模式：`--devices 1`（默认，单设备演示）与 `--devices N>1`（多设备并发压测）。

```bash
# 单设备：暴涨场景（前缓后急冲破红色阈值，最典型）
python sim_water_level.py --scenario surge

# 单设备：指定后端地址、间隔与签名密钥
python sim_water_level.py --api http://127.0.0.1:8080 \
    --device dev-water-001 --scenario rise --interval 2 --secret mysecret

# 只打印水位、不真正上报（无后端时验证曲线）
python sim_water_level.py --scenario wave --dry-run
```

## 多设备并发压测（1000 传感器）

模拟 N 个传感器并发周期上报，输出实时 QPS、成功率与延迟 P50/P95/P99。
脚本会在每个上报间隔内自动打散各设备的触发时刻，避免不真实的瞬时尖峰。

```bash
# 100 传感器、每分钟上报一次、持续运行 5 分钟（贴近生产稳态）
python sim_water_level.py --devices 100 --interval 60 --duration 300 --scenario mixed

# 100 传感器、每设备上报 3 轮后结束（快速验证容量）
python sim_water_level.py --devices 100 --interval 60 --cycles 3

# 仅本地统计、不真正上报（验证脚本与并发模型）
python sim_water_level.py --devices 100 --interval 1 --cycles 2 --dry-run

# 1000 传感器扩展压测（专用脚本，默认即 1000 设备并发）
python sim_water_level_1000.py --interval 60 --duration 300
python sim_water_level_1000.py --interval 10 --duration 120   # 暴雨提频 ≈100 QPS 峰值
```

说明：
- 平均写入 QPS ≈ `devices / interval`。100 设备、60s 间隔 ≈ 1.67 QPS（生产稳态）。
- `--scenario mixed` 会给每个设备随机分配一种曲线，模拟片区内设备状态各异。
- `--duration` 与 `--cycles` 二选一，设了 `--duration` 优先按时长运行。
- `sim_water_level_1000.py` 参数与本脚本完全相同，仅默认值面向 1000 设备。

## 参数说明

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `--api` | `http://127.0.0.1:8080` | 后端基础地址 |
| `--device` | `dev-water-001` | 单设备模式的设备ID |
| `--device-prefix` | `dev-water-` | 多设备模式ID前缀（自动补 4 位序号） |
| `--unit` | `m` | 单位，水位 `m` / 雨量 `mm` |
| `--scenario` | `surge` | 场景：`flat/rise/surge/wave/mixed` |
| `--interval` | `2.0` | 上报间隔（秒） |
| `--devices` | `1` | 模拟设备数，>1 进入并发压测模式 |
| `--workers` | `0` | 并发线程数（0=自动，约 devices/5，上限 500） |
| `--duration` | `0` | 压测运行时长（秒），>0 时优先于 cycles |
| `--cycles` | `3` | 每设备上报轮数（未设 duration 时生效） |
| `--secret` | 空 | 设备密钥，启用 HMAC-SHA256 签名（请求头 `X-Signature`） |
| `--dry-run` | 关闭 | 只统计不上报 |

## 场景曲线

| 场景 | 含义 | 预期效果 |
| --- | --- | --- |
| `flat` | 安全水位小幅波动 | 不触发预警 |
| `rise` | 缓慢线性上涨 | 依次触发 蓝→黄→橙 预警，验证升级 |
| `surge` | 前缓后急暴涨 | 快速冲破红色阈值，验证最高级预警 |
| `wave` | 反复涨落 | 验证防抖（冷却期去重）与升/降级 |
| `mixed` | 每设备随机分配一种曲线 | 多设备压测用，模拟片区设备状态各异 |

## 阈值对照（与 PRD 一致）

| 水位 | 级别 |
| --- | --- |
| ≥ 2.5m | 蓝色 Ⅳ |
| ≥ 3.0m | 黄色 Ⅲ |
| ≥ 3.5m | 橙色 Ⅱ |
| ≥ 4.0m | 红色 Ⅰ |

## 与后端的约定
- 上报体：`{ device_id, value, unit, reported_at }`（ISO8601 UTC 时间）
- 签名（可选）：`X-Signature = HMAC_SHA256(secret, raw_body)` 十六进制小写
- 上报失败自动指数退避重试（最多 3 次）
