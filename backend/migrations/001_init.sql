CREATE TABLE IF NOT EXISTS users (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username    VARCHAR(64) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    display_name VARCHAR(128) NOT NULL DEFAULT '',
    email       VARCHAR(255) NOT NULL DEFAULT '',
    phone       VARCHAR(32)  NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    created_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS roles (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id BIGINT UNSIGNED NOT NULL,
    role_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS permissions (
    id       BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    resource VARCHAR(64) NOT NULL,
    action   VARCHAR(32) NOT NULL,
    UNIQUE KEY uk_resource_action (resource, action)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       BIGINT UNSIGNED NOT NULL,
    permission_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS asset_types (
    id   BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(64) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT IGNORE INTO asset_types (code, name) VALUES
    ('server',  '服务器'),
    ('switch',  '交换机'),
    ('router',  '路由器'),
    ('firewall','防火墙'),
    ('storage', '存储设备'),
    ('vm',      '虚拟机'),
    ('other',   '其他');

CREATE TABLE IF NOT EXISTS assets (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asset_type_id BIGINT UNSIGNED NOT NULL,
    name          VARCHAR(128) NOT NULL,
    ip            VARCHAR(45)  NOT NULL DEFAULT '',
    sn            VARCHAR(128) NOT NULL DEFAULT '',
    manufacturer  VARCHAR(64)  NOT NULL DEFAULT '',
    model         VARCHAR(128) NOT NULL DEFAULT '',
    location      VARCHAR(255) NOT NULL DEFAULT '',
    status        VARCHAR(16)  NOT NULL DEFAULT 'online',
    os_type       VARCHAR(16)  NOT NULL DEFAULT 'linux',
    ssh_user      VARCHAR(64)  NOT NULL DEFAULT '',
    ssh_password  VARCHAR(128) NOT NULL DEFAULT '',
    ssh_port      VARCHAR(8)   NOT NULL DEFAULT '22',
    rdp_user      VARCHAR(64)  NOT NULL DEFAULT '',
    rdp_port      VARCHAR(8)   NOT NULL DEFAULT '3389',
    tags          JSON NOT NULL DEFAULT (JSON_ARRAY()),
    extra         JSON NOT NULL DEFAULT (JSON_OBJECT()),
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (asset_type_id) REFERENCES asset_types(id),
    INDEX idx_ip (ip),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS metric_definitions (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    code        VARCHAR(64) NOT NULL UNIQUE,
    name        VARCHAR(128) NOT NULL,
    unit        VARCHAR(32)  NOT NULL DEFAULT '',
    data_type   VARCHAR(16)  NOT NULL DEFAULT 'gauge',
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS collect_tasks (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asset_id        BIGINT UNSIGNED NOT NULL,
    protocol        VARCHAR(16) NOT NULL DEFAULT 'snmp',
    interval_sec    INT NOT NULL DEFAULT 60,
    enabled         TINYINT NOT NULL DEFAULT 1,
    config          JSON,
    last_collect_at DATETIME,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE,
    INDEX idx_asset (asset_id),
    INDEX idx_enabled (enabled)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alert_rules (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(128) NOT NULL,
    metric_code VARCHAR(64)  NOT NULL,
    condition_op VARCHAR(8)  NOT NULL DEFAULT '>',
    threshold   DOUBLE       NOT NULL DEFAULT 0,
    severity    VARCHAR(16)  NOT NULL DEFAULT 'warning',
    duration    INT          NOT NULL DEFAULT 60,
    enabled     TINYINT      NOT NULL DEFAULT 1,
    extra       JSON,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_metric (metric_code),
    INDEX idx_severity (severity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS alert_events (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    rule_id       BIGINT UNSIGNED NOT NULL,
    asset_id      BIGINT UNSIGNED,
    severity      VARCHAR(16) NOT NULL,
    message       VARCHAR(512) NOT NULL DEFAULT '',
    current_val   DOUBLE       NOT NULL DEFAULT 0,
    status        VARCHAR(16) NOT NULL DEFAULT 'firing',
    fired_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at   DATETIME,
    acked_by      BIGINT UNSIGNED,
    acked_at      DATETIME,
    FOREIGN KEY (rule_id) REFERENCES alert_rules(id),
    FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE SET NULL,
    INDEX idx_status (status),
    INDEX idx_fired (fired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notify_channels (
    id       BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name     VARCHAR(64) NOT NULL,
    type     VARCHAR(16) NOT NULL,
    config   JSON NOT NULL,
    enabled  TINYINT NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS notify_records (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    channel_id  BIGINT UNSIGNED NOT NULL,
    event_id    BIGINT UNSIGNED NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    sent_at     DATETIME,
    error_msg   VARCHAR(512) NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES notify_channels(id),
    FOREIGN KEY (event_id)   REFERENCES alert_events(id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id    BIGINT UNSIGNED,
    action     VARCHAR(64) NOT NULL,
    resource   VARCHAR(64) NOT NULL,
    detail     JSON,
    ip         VARCHAR(45) NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user (user_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS metric_data (
    id           BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    asset_id     BIGINT UNSIGNED NOT NULL,
    metric_code  VARCHAR(64) NOT NULL,
    value        DOUBLE NOT NULL DEFAULT 0,
    collected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_asset_metric (asset_id, metric_code),
    INDEX idx_collected (collected_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


-- Seed data
INSERT INTO users (username, password, display_name, status) VALUES
('admin', '28226046e4b58432d7f1a41a7b8da29d:a671f8e26c710f95ff26c75ad3e4376880e6a0e35eb75f377ef4e4d4fb8e02da', '管理员', 1);

INSERT INTO asset_types (id, name) VALUES
(1, '服务器'), (2, '网络设备'), (3, '存储设备'), (4, '安全设备');

INSERT INTO metric_definitions (code, name, unit) VALUES
('cpu_usage', 'CPU使用率', '%'),
('mem_usage', '内存使用率', '%'),
('disk_usage', '磁盘使用率', '%'),
('load_avg_1m', '1分钟负载', ''),
('net_in', '网络入流量', 'KB/s'),
('net_out', '网络出流量', 'KB/s');

INSERT INTO alert_rules (name, metric_code, condition_op, threshold, severity) VALUES
('CPU使用率警告', 'cpu_usage', '>', 80, 'warning'),
('CPU使用率严重', 'cpu_usage', '>', 95, 'critical'),
('内存使用率警告', 'mem_usage', '>', 80, 'warning'),
('内存使用率严重', 'mem_usage', '>', 95, 'critical'),
('磁盘使用率警告', 'disk_usage', '>', 80, 'warning'),
('磁盘使用率严重', 'disk_usage', '>', 95, 'critical');

INSERT INTO assets (asset_type_id, name, ip, status, os_type, location, ssh_user, ssh_password, ssh_port) VALUES
(1, 'ITSM-Server-Self', 'CHANGE_ME_SERVER_IP', 'online', 'linux', '京东云-北京机房', 'root', 'CHANGE_ME_SSH_PASSWORD', '22');

INSERT INTO assets (asset_type_id, name, ip, status, os_type, location, rdp_user, rdp_port) VALUES
(1, 'WIN-Server', 'CHANGE_ME_WINDOWS_IP', 'online', 'windows', '京东云', 'Administrator', '3389');
