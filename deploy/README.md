# 部署指南

ITSM_ops 支持三种部署方式，共用同一份 Docker 镜像。

---

## 一、本地开发（无 Docker）

### 环境要求

- Go 1.24+、gcc（cgo 编译 SQLite 驱动）
- Node.js 24+ / npm 11+
- SQLite 3

### 步骤

```bash
# 1. 初始化数据库
mkdir -p data
sqlite3 data/itsm.db < backend/migrations/001_init_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/002_users_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/003_metrics_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/004_alerts_sqlite.sql
cd backend && CGO_ENABLED=1 go run ./cmd/seed/main.go | sqlite3 ../data/itsm.db && cd ..

# 2. 启动后端 (端口 8080)
cd backend
DB_TYPE=sqlite3 DB_NAME=../data/itsm.db JWT_SECRET=dev-secret \
  CGO_ENABLED=1 go run ./cmd/server/main.go

# 3. 启动前端 (端口 3000，另开终端)
cd frontend && npm install && npm run dev

# 4. 启动采集器 (另开终端，可选)
cd collector
COLLECTOR_TOKEN=<JWT> COLLECTOR_ASSET_ID=1 go run ./cmd/agent/main.go

# 5. 访问 http://localhost:3000  账号: admin / admin@123
```

---

## 二、Docker Compose

```bash
cd deploy/compose
cp .env.example .env    # 修改 JWT_SECRET 和数据库密码
docker compose up -d --build
```

| 服务 | 端口 |
| --- | --- |
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080 |

生产覆盖（端口 80 + restart always）：

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

### 镜像列表

| 镜像 | Dockerfile | 基础镜像 |
| --- | --- | --- |
| itsm-ops/backend | deploy/docker/backend.Dockerfile | golang:1.24-alpine → alpine:3.21 |
| itsm-ops/frontend | deploy/docker/frontend.Dockerfile | node:24-alpine → nginx:1.27-alpine |
| itsm-ops/collector | deploy/docker/collector.Dockerfile | golang:1.24-alpine → alpine:3.21 |

---

## 三、Kubernetes

### Kustomize

```bash
kubectl apply -k deploy/k8s/overlays/dev/       # 开发（单副本）
kubectl apply -k deploy/k8s/overlays/staging/    # Staging（2 副本）
kubectl apply -k deploy/k8s/overlays/prod/       # 生产（3 副本）
```

### Helm Chart

```bash
helm install itsm deploy/k8s/charts/itsm-ops/ \
  --set secrets.jwtSecret=your-jwt-secret \
  --set secrets.dbPassword=your-db-password \
  --set ingress.host=itsm.your-domain.com
```

### K8s 资源

| 资源 | 名称 | 说明 |
| --- | --- | --- |
| Deployment | itsm-backend | API + 告警引擎 |
| Deployment | itsm-frontend | Nginx 静态文件 |
| DaemonSet | itsm-collector | 每节点采集器 |
| Service | itsm-backend / frontend | ClusterIP |
| ConfigMap | itsm-config | 非敏感配置 |
| Secret | itsm-secret | 密码 / JWT 密钥 |
| Ingress | itsm-ingress | /api → backend, / → frontend |

---

## 四、配置参考

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| APP_ENV | 运行环境 | development |
| APP_PORT | HTTP 端口 | 8080 |
| DB_TYPE | sqlite3 / mysql | sqlite3 |
| DB_HOST | 数据库地址 | 127.0.0.1 |
| DB_PORT | 数据库端口 | 3306 |
| DB_NAME | 库名 / SQLite 路径 | itsm_ops.db |
| DB_USER / DB_PASSWORD | 数据库认证 | itsm / (空) |
| JWT_SECRET | JWT 签名密钥 | please-change-me |
| JWT_EXPIRE | JWT 过期时间 | 24h |

---

## 五、目录结构

```
deploy/
├── docker/
│   ├── backend.Dockerfile
│   ├── frontend.Dockerfile
│   ├── collector.Dockerfile
│   ├── nginx.conf              HTTP 配置
│   └── nginx-tls.conf          HTTPS 配置
├── compose/
│   ├── docker-compose.yml
│   ├── docker-compose.prod.yml
│   └── .env.example
└── k8s/
    ├── base/                   Kustomize 基线 (8 资源)
    ├── overlays/dev|staging|prod
    └── charts/itsm-ops/        Helm Chart (7 模板)
```
