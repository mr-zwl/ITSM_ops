CREATE TABLE IF NOT EXISTS asset_types (
    id   INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL
);

INSERT OR IGNORE INTO asset_types (code, name) VALUES
    ('server',  '服务器'),
    ('switch',  '交换机'),
    ('router',  '路由器'),
    ('firewall','防火墙'),
    ('storage', '存储设备'),
    ('vm',      '虚拟机'),
    ('other',   '其他');

CREATE TABLE IF NOT EXISTS assets (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    asset_type_id INTEGER NOT NULL REFERENCES asset_types(id),
    name          TEXT NOT NULL,
    ip            TEXT NOT NULL DEFAULT '',
    sn            TEXT NOT NULL DEFAULT '',
    manufacturer  TEXT NOT NULL DEFAULT '',
    model         TEXT NOT NULL DEFAULT '',
    location      TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'online',
    tags          TEXT DEFAULT '[]',
    extra         TEXT DEFAULT '{}',
    created_at    DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at    DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_assets_ip ON assets(ip);
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
