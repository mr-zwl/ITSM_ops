# Ansible 物理机自动化部署

一键将 ITSM_ops 平台部署到物理机 / 虚拟机，部署工程师**只需修改 hosts.ini 和 all.yml**。

---

## 快速开始

### 1. 构建发布包

```bash
cd /path/to/ITSM_ops
bash scripts/build/build-all.sh
```

构建产物输出到 `dist/` 目录：

```
dist/
├── itsm-ops-backend     后端二进制
├── itsm-seed            管理员用户 seed 工具
├── itsm-collector       采集器二进制
└── frontend-dist/       前端静态文件
```

### 2. 修改清单文件

编辑 `deploy/ansible/inventories/prod/hosts.ini`：

```ini
[backend]
10.0.1.10 ansible_user=root

[frontend]
10.0.1.20 ansible_user=root

[collector]
10.0.1.10 ansible_user=root
10.0.1.30 ansible_user=root
```

### 3. 修改配置变量

编辑 `deploy/ansible/group_vars/all.yml`：

```yaml
# 必须修改
db_password: "your-real-db-password"
jwt_secret: "your-real-jwt-secret"
collector_token: "your-collector-jwt-token"
frontend_domain: itsm.your-company.com

# 按需修改
db_type: mysql           # mysql 或 sqlite3
db_host: 10.0.1.100     # MySQL 地址
```

### 4. 执行部署

```bash
cd deploy/ansible

# 全量部署
ansible-playbook site.yml

# 只部署后端
ansible-playbook site.yml --tags backend

# 只部署某台机器
ansible-playbook site.yml --limit 10.0.1.10

# 试运行（不实际执行）
ansible-playbook site.yml --check
```

---

## 部署架构

```
                    ┌──────────────┐
                    │  浏览器访问   │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
 [frontend]         │  Nginx       │
 10.0.1.20          │  静态文件    │
                    │  反向代理    │──── /api/ ────┐
                    └──────────────┘               │
                                           ┌──────▼───────┐
 [backend]                                 │  Backend     │
 10.0.1.10                                 │  :8080       │
                                           │  告警引擎    │
                                           └──────┬───────┘
                                                  │
 [collector]      ┌──────────────┐                │
 10.0.1.10        │  Collector   │── 上报指标 ────┘
 10.0.1.30        │  Collector   │── 上报指标 ────┘
                  └──────────────┘
```

---

## 部署后验证

```bash
# 健康检查
curl http://backend-ip:8080/healthz

# 登录测试
curl -X POST -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin@123"}' \
  http://backend-ip:8080/api/v1/auth/login

# 前端页面
curl http://frontend-ip/
```

---

## 目录结构

```
deploy/ansible/
├── ansible.cfg                     Ansible 配置
├── site.yml                        主 Playbook（入口）
├── inventories/
│   └── prod/
│       └── hosts.ini               主机清单（部署工程师修改此文件）
├── group_vars/
│   └── all.yml                     全局变量（部署工程师修改此文件）
└── roles/
    ├── common/tasks/main.yml       系统用户 + 目录 + 依赖
    ├── backend/
    │   ├── tasks/main.yml          二进制部署 + 迁移 + systemd
    │   ├── templates/
    │   │   ├── backend.env.j2      环境变量文件
    │   │   └── itsm-ops-backend.service.j2
    │   └── handlers/main.yml
    ├── frontend/
    │   ├── tasks/main.yml          Nginx + 静态文件部署
    │   ├── templates/
    │   │   └── itsm-ops.conf.j2    Nginx 配置（含 SSL 支持）
    │   └── handlers/main.yml
    └── collector/
        ├── tasks/main.yml          采集器二进制 + systemd
        ├── templates/
        │   ├── collector.env.j2
        │   └── itsm-ops-collector.service.j2
        └── handlers/main.yml
```

---

## 服务管理

部署后可以直接用 systemd 管理：

```bash
# 后端
systemctl status itsm-ops-backend
systemctl restart itsm-ops-backend
journalctl -u itsm-ops-backend -f

# 采集器
systemctl status itsm-ops-collector
systemctl restart itsm-ops-collector

# 前端 (Nginx)
systemctl status nginx
nginx -t && systemctl reload nginx
```

日志位置：`/opt/itsm-ops/logs/`

---

## 启用 HTTPS

在 `group_vars/all.yml` 中设置：

```yaml
frontend_ssl: true
frontend_ssl_cert: /etc/nginx/ssl/itsm.crt
frontend_ssl_key: /etc/nginx/ssl/itsm.key
```

提前将证书放到前端服务器对应路径，重新部署即可。
