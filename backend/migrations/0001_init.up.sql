-- 0001_init: 核心表结构（PostgreSQL）
-- 角色：villager / gridworker（网格员兼后台管理）

CREATE TABLE IF NOT EXISTS grids (
    id              BIGSERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    village         TEXT NOT NULL DEFAULT '',
    manager_user_id BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    openid     TEXT NOT NULL UNIQUE,
    role       TEXT NOT NULL DEFAULT 'villager',
    name       TEXT NOT NULL DEFAULT '',
    phone      TEXT NOT NULL DEFAULT '',
    grid_id    BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS devices (
    id             TEXT PRIMARY KEY,
    type           TEXT NOT NULL DEFAULT 'water_level',
    name           TEXT NOT NULL DEFAULT '',
    grid_id        BIGINT NOT NULL DEFAULT 0,
    status         TEXT NOT NULL DEFAULT 'online',
    last_value     NUMERIC(8,3) NOT NULL DEFAULT 0,
    last_report_at TIMESTAMPTZ
);

-- 读数时序大表：按天范围分区
CREATE TABLE IF NOT EXISTS readings (
    id          BIGINT GENERATED ALWAYS AS IDENTITY,
    device_id   TEXT NOT NULL,
    value       NUMERIC(8,3) NOT NULL,
    unit        TEXT NOT NULL DEFAULT 'm',
    reported_at TIMESTAMPTZ NOT NULL
) PARTITION BY RANGE (reported_at);

CREATE INDEX IF NOT EXISTS idx_readings_brin ON readings USING BRIN (reported_at);
CREATE INDEX IF NOT EXISTS idx_readings_dev ON readings (device_id, reported_at DESC);

-- 兜底默认分区，避免无分区时插入失败（cron 会创建按天分区）
CREATE TABLE IF NOT EXISTS readings_default PARTITION OF readings DEFAULT;

-- 日聚合趋势表（降采样，永久保留）
CREATE TABLE IF NOT EXISTS daily_water_stats (
    device_id  TEXT NOT NULL,
    metric     TEXT NOT NULL,
    stat_date  DATE NOT NULL,
    max_value  NUMERIC(8,3) NOT NULL,
    avg_value  NUMERIC(8,3) NOT NULL,
    min_value  NUMERIC(8,3),
    sample_cnt INTEGER NOT NULL,
    unit       TEXT,
    PRIMARY KEY (device_id, metric, stat_date)
);

CREATE TABLE IF NOT EXISTS templates (
    id            BIGSERIAL PRIMARY KEY,
    name          TEXT NOT NULL,
    disaster_type TEXT NOT NULL DEFAULT 'flood',
    content_tpl   TEXT NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS rules (
    id              BIGSERIAL PRIMARY KEY,
    device_type     TEXT NOT NULL DEFAULT 'water_level',
    metric          TEXT NOT NULL DEFAULT 'water_level',
    operator        TEXT NOT NULL DEFAULT '>=',
    threshold       NUMERIC(8,3) NOT NULL,
    level           TEXT NOT NULL,
    cooldown_sec    INTEGER NOT NULL DEFAULT 600,
    template_id     BIGINT NOT NULL DEFAULT 0,
    review_required BOOLEAN NOT NULL DEFAULT false,
    enabled         BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS alerts (
    id            BIGSERIAL PRIMARY KEY,
    source        TEXT NOT NULL DEFAULT 'sensor',
    level         TEXT NOT NULL,
    disaster_type TEXT NOT NULL DEFAULT 'flood',
    grid_id       BIGINT NOT NULL DEFAULT 0,
    device_id     TEXT NOT NULL DEFAULT '',
    title         TEXT NOT NULL DEFAULT '',
    content       TEXT NOT NULL DEFAULT '',
    tts_url       TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'triggered',
    triggered_by  BIGINT NOT NULL DEFAULT 0,
    reviewed_by   BIGINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts (status);
CREATE INDEX IF NOT EXISTS idx_alerts_grid ON alerts (grid_id, created_at DESC);

CREATE TABLE IF NOT EXISTS alert_logs (
    id          BIGSERIAL PRIMARY KEY,
    alert_id    BIGINT NOT NULL,
    from_status TEXT NOT NULL DEFAULT '',
    to_status   TEXT NOT NULL,
    operator_id BIGINT NOT NULL DEFAULT 0,
    remark      TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_alert_logs_alert ON alert_logs (alert_id, created_at);

CREATE TABLE IF NOT EXISTS tasks (
    id            BIGSERIAL PRIMARY KEY,
    alert_id      BIGINT NOT NULL,
    assignee_id   BIGINT NOT NULL DEFAULT 0,
    status        TEXT NOT NULL DEFAULT 'pending',
    handle_remark TEXT NOT NULL DEFAULT '',
    finished_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks (assignee_id, status);

CREATE TABLE IF NOT EXISTS attachments (
    id         BIGSERIAL PRIMARY KEY,
    task_id    BIGINT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'image',
    cos_key    TEXT NOT NULL DEFAULT '',
    url        TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reports (
    id          BIGSERIAL PRIMARY KEY,
    reporter_id BIGINT NOT NULL DEFAULT 0,
    grid_id     BIGINT NOT NULL DEFAULT 0,
    content     TEXT NOT NULL DEFAULT '',
    lng         NUMERIC(10,6) NOT NULL DEFAULT 0,
    lat         NUMERIC(10,6) NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'open',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS report_attachments (
    id        BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL,
    type      TEXT NOT NULL DEFAULT 'image',
    cos_key   TEXT NOT NULL DEFAULT '',
    url       TEXT NOT NULL DEFAULT ''
);
