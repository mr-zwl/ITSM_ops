CREATE TABLE IF NOT EXISTS alert_rules (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT NOT NULL,
    metric_code  TEXT NOT NULL,
    condition_op TEXT NOT NULL DEFAULT '>',
    threshold    REAL NOT NULL DEFAULT 0,
    severity     TEXT NOT NULL DEFAULT 'warning',
    duration_sec INTEGER NOT NULL DEFAULT 0,
    enabled      INTEGER NOT NULL DEFAULT 1,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO alert_rules (name, metric_code, condition_op, threshold, severity) VALUES
    ('CPU 使用率过高',  'cpu_usage',  '>', 80, 'warning'),
    ('CPU 使用率严重',  'cpu_usage',  '>', 95, 'critical'),
    ('内存使用率过高',  'mem_usage',  '>', 85, 'warning'),
    ('内存使用率严重',  'mem_usage',  '>', 95, 'critical'),
    ('磁盘使用率过高',  'disk_usage', '>', 80, 'warning'),
    ('磁盘使用率严重',  'disk_usage', '>', 95, 'critical');

CREATE TABLE IF NOT EXISTS alert_events (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    rule_id     INTEGER NOT NULL REFERENCES alert_rules(id),
    asset_id    INTEGER REFERENCES assets(id) ON DELETE SET NULL,
    severity    TEXT NOT NULL,
    message     TEXT NOT NULL DEFAULT '',
    current_val REAL NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'firing',
    fired_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    resolved_at DATETIME,
    acked_by    INTEGER,
    acked_at    DATETIME
);

CREATE INDEX IF NOT EXISTS idx_alert_events_status ON alert_events(status);
CREATE INDEX IF NOT EXISTS idx_alert_events_fired ON alert_events(fired_at);

CREATE TABLE IF NOT EXISTS notify_channels (
    id       INTEGER PRIMARY KEY AUTOINCREMENT,
    name     TEXT NOT NULL,
    type     TEXT NOT NULL DEFAULT 'log',
    config   TEXT NOT NULL DEFAULT '{}',
    enabled  INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO notify_channels (name, type, config) VALUES
    ('控制台日志', 'log', '{}');

CREATE TABLE IF NOT EXISTS notify_records (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    channel_id INTEGER NOT NULL REFERENCES notify_channels(id),
    event_id   INTEGER NOT NULL REFERENCES alert_events(id),
    status     TEXT NOT NULL DEFAULT 'sent',
    message    TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now'))
);
