CREATE TABLE IF NOT EXISTS metric_definitions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    unit        TEXT NOT NULL DEFAULT '',
    data_type   TEXT NOT NULL DEFAULT 'gauge',
    description TEXT NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO metric_definitions (code, name, unit, data_type) VALUES
    ('cpu_usage',     'CPU 使用率',   '%',   'gauge'),
    ('mem_usage',     '内存使用率',   '%',   'gauge'),
    ('disk_usage',    '磁盘使用率',   '%',   'gauge'),
    ('net_in_bytes',  '网络入流量',   'B/s', 'counter'),
    ('net_out_bytes', '网络出流量',   'B/s', 'counter'),
    ('load_avg_1m',   '1分钟负载',    '',    'gauge'),
    ('uptime_sec',    '运行时长',     's',   'counter');

CREATE TABLE IF NOT EXISTS metric_data (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id   INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    metric_code TEXT NOT NULL,
    value      REAL NOT NULL,
    collected_at DATETIME NOT NULL DEFAULT (datetime('now')),
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_metric_data_asset ON metric_data(asset_id);
CREATE INDEX IF NOT EXISTS idx_metric_data_code ON metric_data(metric_code);
CREATE INDEX IF NOT EXISTS idx_metric_data_collected ON metric_data(collected_at);

CREATE TABLE IF NOT EXISTS collect_tasks (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_id       INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    protocol       TEXT NOT NULL DEFAULT 'ssh',
    interval_sec   INTEGER NOT NULL DEFAULT 60,
    enabled        INTEGER NOT NULL DEFAULT 1,
    config         TEXT DEFAULT '{}',
    last_collect_at DATETIME,
    created_at     DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at     DATETIME NOT NULL DEFAULT (datetime('now'))
);
