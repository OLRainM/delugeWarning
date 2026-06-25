#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
水位/雨量传感器数据模拟脚本 —— 1000 传感器压测版本
=========================================
用途：模拟乡村暴雨夜场景下，水位传感器向后端定时上报读数，
     从而驱动后端规则引擎"系统自动触发预警"，便于联调与压测。

本版本面向 **1000 传感器** 规模，默认进入多设备并发压测模式。
（100 传感器规模请使用 sim_water_level.py）

两种模式：
  1) 单设备模式（--devices 1）：直观演示单个设备的水位曲线变化。
  2) 多设备压测模式（--devices N，默认 1000）：模拟 N 个传感器并发周期上报，
     用于评估后端在 1000 传感器规模下的 QPS、成功率与延迟表现。

特性：
  - 多种水位变化曲线：平稳(flat)、缓涨(rise)、暴涨(surge)、涨落(wave)、混合(mixed)
  - 多设备并发上报（线程池），可设运行时长或上报轮数
  - 在上报间隔内自动打散各设备触发时刻，避免不真实的瞬时尖峰
  - 实时统计：实际 QPS、成功/失败、延迟 P50/P95/P99
  - 支持 HMAC 签名上报（与后端设备签名校验对齐，可关闭）
  - 上报失败自动重试，纯标准库实现，无第三方依赖

用法示例：
  # 1000 传感器、每分钟上报一次、运行 5 分钟的压测（生产稳态）
  python sim_water_level_1000.py --interval 60 --duration 300 --scenario mixed
  # 1000 传感器、每设备上报 3 轮后结束（快速验证）
  python sim_water_level_1000.py --cycles 3 --workers 200
  # 暴雨提频压测：每 10 秒一次（≈100 QPS 峰值）
  python sim_water_level_1000.py --interval 10 --duration 120
  # 只统计不上报
  python sim_water_level_1000.py --dry-run
"""

import argparse
import hashlib
import hmac
import json
import math
import random
import threading
import time
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone


# ---------- 水位曲线场景定义 ----------
# 返回一个生成器，逐步产出水位值(米)。结合阈值(蓝2.5/黄3.0/橙3.5/红4.0)设计。
def scenario_flat(start=2.0, steps=60):
    """平稳：在安全水位附近小幅波动，不触发预警。"""
    for _ in range(steps):
        yield round(start + random.uniform(-0.05, 0.05), 3)


def scenario_rise(start=2.0, peak=3.6, steps=60):
    """缓涨：水位线性缓慢上升，依次穿过蓝/黄/橙阈值。"""
    for i in range(steps):
        base = start + (peak - start) * (i / max(steps - 1, 1))
        yield round(base + random.uniform(-0.03, 0.03), 3)


def scenario_surge(start=2.2, peak=4.3, steps=40):
    """暴涨：前期平稳，中段急剧上涨冲破红色阈值（暴雨夜典型）。"""
    for i in range(steps):
        ratio = i / max(steps - 1, 1)
        # 使用 S 形曲线，前缓后急
        s = 1 / (1 + math.exp(-12 * (ratio - 0.55)))
        base = start + (peak - start) * s
        yield round(base + random.uniform(-0.04, 0.04), 3)


def scenario_wave(start=2.0, peak=3.7, steps=80):
    """涨落：水位反复涨落，验证防抖与升级/降级逻辑。"""
    for i in range(steps):
        base = start + (peak - start) * (0.5 - 0.5 * math.cos(i / 6.0))
        yield round(base + random.uniform(-0.05, 0.05), 3)


SCENARIOS = {
    "flat": scenario_flat,
    "rise": scenario_rise,
    "surge": scenario_surge,
    "wave": scenario_wave,
}


def make_value_provider(scenario: str):
    """为单个设备返回一个"取下一个水位值"的函数。

    - 普通场景：复用生成器，耗尽后稳定在安全水位附近，保证压测可长期运行。
    - mixed：按设备随机分配一种场景，模拟真实片区中设备状态各异。
    """
    name = scenario
    if scenario == "mixed":
        name = random.choice(list(SCENARIOS.keys()))
    gen = SCENARIOS[name]()

    def _next() -> float:
        try:
            return next(gen)
        except StopIteration:
            return round(2.0 + random.uniform(-0.05, 0.05), 3)

    return name, _next


LEVELS = [(4.0, "红色/Ⅰ"), (3.5, "橙色/Ⅱ"), (3.0, "黄色/Ⅲ"), (2.5, "蓝色/Ⅳ")]


def level_of(value: float) -> str:
    for thr, name in LEVELS:
        if value >= thr:
            return name
    return "正常"


def build_payload(device_id: str, value: float, unit: str) -> dict:
    return {
        "device_id": device_id,
        "value": value,
        "unit": unit,
        "reported_at": datetime.now(timezone.utc).isoformat(),
    }


def sign(body: bytes, secret: str) -> str:
    """与后端设备签名校验对齐：HMAC-SHA256(secret, body) 的十六进制。"""
    return hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()


def post_reading(api: str, device_id: str, payload: dict, secret: str = "",
                 retries: int = 3, timeout: int = 5, verbose: bool = True):
    """上报一条读数。返回 (是否成功, 耗时毫秒)。"""
    url = f"{api.rstrip('/')}/api/v1/devices/{device_id}/readings"
    body = json.dumps(payload).encode("utf-8")
    headers = {"Content-Type": "application/json"}
    if secret:
        headers["X-Signature"] = sign(body, secret)

    start = time.perf_counter()
    for attempt in range(1, retries + 1):
        try:
            req = urllib.request.Request(url, data=body, headers=headers, method="POST")
            with urllib.request.urlopen(req, timeout=timeout) as resp:
                ok = 200 <= resp.status < 300
                return ok, (time.perf_counter() - start) * 1000
        except urllib.error.HTTPError as e:
            if verbose:
                print(f"  [上报失败] {device_id} HTTP {e.code}: {e.reason} (第{attempt}次)")
        except Exception as e:  # noqa: BLE001
            if verbose:
                print(f"  [上报异常] {device_id} {e} (第{attempt}次)")
        time.sleep(min(2 ** attempt, 5))
    return False, (time.perf_counter() - start) * 1000


class Stats:
    """线程安全的上报统计聚合器。"""

    def __init__(self):
        self._lock = threading.Lock()
        self.ok = 0
        self.fail = 0
        self.alerts = 0
        self.latencies = []  # 毫秒

    def record(self, success: bool, latency_ms: float, is_alert: bool):
        with self._lock:
            if success:
                self.ok += 1
            else:
                self.fail += 1
            if is_alert:
                self.alerts += 1
            self.latencies.append(latency_ms)

    def snapshot(self):
        with self._lock:
            return self.ok, self.fail, self.alerts, list(self.latencies)


def _percentile(data, p):
    if not data:
        return 0.0
    s = sorted(data)
    k = max(0, min(len(s) - 1, int(round((p / 100.0) * (len(s) - 1)))))
    return s[k]


def run_single(args):
    """单设备模式：直观打印水位曲线变化。"""
    name, next_val = make_value_provider(args.scenario)
    print(f"开始模拟[单设备]：设备={args.device} 场景={name} "
          f"间隔={args.interval}s 目标={args.api} dry_run={args.dry_run}")
    print("-" * 64)
    ok_count = total = 0
    # 普通场景跑完曲线即结束；mixed 等无界场景按 cycles 控制条数
    max_steps = None if args.scenario in SCENARIOS else max(args.cycles, 1) * 40
    try:
        while max_steps is None or total < max_steps:
            value = round(next_val(), 3)
            total += 1
            lvl = level_of(value)
            ts = datetime.now().strftime("%H:%M:%S")
            flag = "⚠️ 预警级别" if lvl != "正常" else "正常"
            print(f"[{ts}] 水位={value}{args.unit}  -> {flag}: {lvl}")
            if not args.dry_run:
                payload = build_payload(args.device, value, args.unit)
                ok, _ = post_reading(args.api, args.device, payload, args.secret)
                ok_count += 1 if ok else 0
                if not ok:
                    print("  -> 上报最终失败，跳过本条")
            # 普通场景曲线耗尽（回到安全水位附近且已跑足）则结束
            if args.scenario in SCENARIOS and total >= _scenario_steps(args.scenario):
                break
            time.sleep(args.interval)
    except KeyboardInterrupt:
        print("\n已手动中断。")
    print("-" * 64)
    print(f"结束：共生成 {total} 条" + ("" if args.dry_run else f"，成功上报 {ok_count} 条"))


def _scenario_steps(scenario: str) -> int:
    """各内置场景生成器的步数，用于单设备模式判断何时跑完一条曲线。"""
    return {"flat": 60, "rise": 60, "surge": 40, "wave": 80}.get(scenario, 60)


def run_multi(args):
    """多设备压测模式：N 个传感器并发周期上报。"""
    n = args.devices
    devices = [(f"{args.device_prefix}{i:04d}", make_value_provider(args.scenario)[1])
               for i in range(1, n + 1)]
    stats = Stats()
    stop_at = time.time() + args.duration if args.duration > 0 else None
    workers = args.workers or min(500, max(50, n // 5))
    print(f"开始压测[多设备]：设备数={n} 场景={args.scenario} 间隔={args.interval}s "
          f"并发={workers} 目标={args.api}")
    print(f"目标平均写入 QPS ≈ {n / args.interval:.1f}  "
          + (f"运行{args.duration}s" if stop_at else f"每设备{args.cycles}轮") + " dry_run="
          + str(args.dry_run))
    print("-" * 64)

    def task(device_id, next_val):
        value = round(next_val(), 3)
        is_alert = level_of(value) != "正常"
        if args.dry_run:
            stats.record(True, 0.0, is_alert)
            return
        payload = build_payload(device_id, value, args.unit)
        ok, ms = post_reading(args.api, device_id, payload, args.secret, verbose=False)
        stats.record(ok, ms, is_alert)

    t0 = time.time()
    cycle = 0
    try:
        with ThreadPoolExecutor(max_workers=workers) as pool:
            while True:
                cycle += 1
                cycle_start = time.time()
                # 在 interval 内打散每个设备的触发时刻，避免不真实的瞬时尖峰
                for idx, (dev, nv) in enumerate(devices):
                    jitter = (idx / n) * args.interval if n > 0 else 0
                    pool.submit(_delayed_submit, task, dev, nv, jitter)
                ok, fail, alerts, lat = stats.snapshot()
                done = ok + fail
                elapsed = max(time.time() - t0, 1e-6)
                print(f"[轮次{cycle}] 已上报={done} 成功={ok} 失败={fail} 预警样本={alerts} "
                      f"实时QPS={done / elapsed:.1f} "
                      f"P50={_percentile(lat, 50):.0f}ms P95={_percentile(lat, 95):.0f}ms "
                      f"P99={_percentile(lat, 99):.0f}ms")
                if stop_at and time.time() >= stop_at:
                    break
                if not stop_at and cycle >= args.cycles:
                    break
                sleep_left = args.interval - (time.time() - cycle_start)
                if sleep_left > 0:
                    time.sleep(sleep_left)
    except KeyboardInterrupt:
        print("\n已手动中断，等待在途请求结束...")

    ok, fail, alerts, lat = stats.snapshot()
    dur = max(time.time() - t0, 1e-6)
    print("-" * 64)
    print(f"压测结束：时长={dur:.1f}s 总请求={ok + fail} 成功={ok} 失败={fail} "
          f"预警样本={alerts}")
    print(f"平均QPS={ (ok + fail) / dur:.1f}  成功率={ (ok / (ok + fail) * 100) if (ok+fail) else 0:.1f}%")
    print(f"延迟(ms): P50={_percentile(lat,50):.0f} P95={_percentile(lat,95):.0f} "
          f"P99={_percentile(lat,99):.0f} max={max(lat) if lat else 0:.0f}")


def _delayed_submit(task, dev, nv, jitter):
    if jitter > 0:
        time.sleep(jitter)
    task(dev, nv)


def main():
    parser = argparse.ArgumentParser(
        description="水位/雨量传感器数据模拟脚本（1000 传感器压测版）")
    parser.add_argument("--api", default="http://127.0.0.1:8080", help="后端基础地址")
    parser.add_argument("--device", default="dev-water-0001", help="单设备模式的设备ID")
    parser.add_argument("--device-prefix", default="dev-water-", help="多设备模式ID前缀")
    parser.add_argument("--unit", default="m", help="读数单位，水位 m / 雨量 mm")
    parser.add_argument("--scenario", default="mixed",
                        choices=list(SCENARIOS.keys()) + ["mixed"],
                        help="水位场景：flat/rise/surge/wave/mixed（默认 mixed）")
    parser.add_argument("--interval", type=float, default=60.0, help="上报间隔(秒)，默认60")
    parser.add_argument("--devices", type=int, default=1000, help="模拟设备数，默认1000")
    parser.add_argument("--workers", type=int, default=0, help="并发线程数(0=自动)")
    parser.add_argument("--duration", type=float, default=0, help="压测运行时长(秒)，>0 优先")
    parser.add_argument("--cycles", type=int, default=3, help="每设备上报轮数(未设duration时生效)")
    parser.add_argument("--secret", default="", help="设备密钥(HMAC签名)，为空则不签名")
    parser.add_argument("--dry-run", action="store_true", help="只统计不真正上报")
    args = parser.parse_args()

    if args.devices > 1:
        run_multi(args)
    else:
        run_single(args)


if __name__ == "__main__":
    main()
