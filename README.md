# ITSM_ops · Web 运维智能监控一体化平台

> 面向复杂 IT 架构的运维监控平台 —— 统一管理服务器、交换机、存储设备，自动采集监控指标，内建告警引擎与通知系统。

---

## 技术栈

| 层 | 技术 |
| --- | --- |
| 后端 | Go 1.24 · net/http · sqlx · SQLite/MySQL |
| 前端 | Vue 3 · TypeScript · Vite · Pinia · Vue Router |
| 采集器 | Go · Linux 系统命令 (CPU/内存/磁盘/负载) |
| 部署 | Docker Compose · Kubernetes (Kustomize + Helm) |
| CI/CD | GitHub Actions |

---

## 快速开始

### 环境要求

- Go 1.24+（后端/采集器编译需要 gcc）
- Node.js 24+ / npm 11+
- SQLite 3（开发环境）或 MySQL 5.7+（生产环境）

### 1. 克隆项目

```bash
git clone https://github.com/your-org/ITSM_ops.git
cd ITSM_ops
```

### 2. 初始化数据库

```bash
sqlite3 data/itsm.db < backend/migrations/001_init_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/002_users_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/003_metrics_sqlite.sql
sqlite3 data/itsm.db < backend/migrations/004_alerts_sqlite.sql

cd backend && CGO_ENABLED=1 go run ./cmd/seed/main.go | sqlite3 ../data/itsm.db && cd ..
```

### 3. 启动后端

```bash
cd backend
DB_TYPE=sqlite3 DB_NAME=../data/itsm.db JWT_SECRET=your-secret-key \
  CGO_ENABLED=1 go run ./cmd/server/main.go
```

后端默认监听 `:8080`，告警引擎每 30 秒自动评估一次。

### 4. 启动前端

```bash
cd frontend
npm install
npm run dev
```

前端默认监听 `:3000`，Vite 自动代理 `/api` 到后端 `localhost:8080`。

### 5. 启动采集器（可选）

```bash
cd collector

COLLECTOR_ENDPOINT=http://127.0.0.1:8080/api/v1/collect \
COLLECTOR_TOKEN=<登录后获取的JWT> \
COLLECTOR_ASSET_ID=1 \
COLLECTOR_INTERVAL=30 \
  go run ./cmd/agent/main.go
```

### 6. 访问

打开 `http://localhost:3000`，默认账号：

```
用户名：admin
密码：  admin@123
```

---

## Docker Compose 部署

```bash
cd deploy/compose
cp .env.example .env    # 修改密码和密钥
docker compose up -d --build
```

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080 |
| 健康检查 | http://localhost:8080/healthz |

生产环境加载覆盖文件：

```bash
docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build
```

---

## Kubernetes 部署

### Kustomize

```bash
# 开发环境（单副本）
kubectl apply -k deploy/k8s/overlays/dev/

# 生产环境（3 副本 + Ingress）
kubectl apply -k deploy/k8s/overlays/prod/
```

### Helm

```bash
helm install itsm deploy/k8s/charts/itsm-ops/ \
  --set secrets.jwtSecret=your-secret \
  --set secrets.dbPassword=your-db-password \
  --set ingress.host=itsm.your-domain.com
```

---

## 功能现状

### 已实现

| 模块 | 能力 |
| --- | --- |
| 认证 | JWT 登录、中间件鉴权、密码哈希 |
| 资产管理 | 服务器/交换机/存储 CRUD、7 种资产类型 |
| 指标采集 | 采集器 agent（Linux real + mock）、指标上报、指标查询 |
| 告警引擎 | 6 条内置规则、阈值评估（30s 周期）、告警事件管理 |
| 通知系统 | 控制台日志渠道、通知记录 |
| 前端 | 登录、仪表盘、资产管理（增删查改）、指标监控（CSS 仪表）、告警事件（确认操作） |
| 部署 | Docker Compose、Kubernetes（Kustomize 三环境 + Helm Chart） |
| CI/CD | GitHub Actions（构建/测试/镜像推送） |
| 安全 | IP 限流中间件（100 次/分）、HTTPS nginx 配置 |
| 测试 | JWT/密码/配置/限流 单元测试 |
| API 文档 | OpenAPI 3.0.3 规范 |

### 待实现（路线图）

- [ ] 拓扑自动发现（LLDP/CDP）
- [ ] 可视化大屏（ECharts）
- [ ] 多租户数据隔离
- [ ] RBAC 精细化（按资产组/区域）
- [ ] AI 异常检测（历史基线）
- [ ] ITSM 工单集成
- [ ] 国际化 i18n

---

## 目录结构

```
ITSM_ops/
├── backend/                 Go 后端服务
│   ├── cmd/server/          主入口
│   ├── cmd/seed/            管理员用户 seed 工具
│   ├── internal/
│   │   ├── api/             路由 + 响应
│   │   ├── asset/           资产 CRUD (handler + repository)
│   │   ├── auth/            JWT 认证 (handler + middleware)
│   │   ├── metric/          指标管理 + 数据上报
│   │   ├── alert/           告警引擎 + 事件管理
│   │   ├── notify/          通知发送 + 记录
│   │   ├── middleware/      限流中间件
│   │   ├── config/          环境变量配置
│   │   ├── topology/        拓扑（桩）
│   │   └── report/          报表（桩）
│   ├── pkg/
│   │   ├── auth/            JWT 签发/验证 + 密码哈希
│   │   ├── db/              数据库连接池 (SQLite/MySQL)
│   │   └── logger/          结构化日志 (slog)
│   └── migrations/          SQL 迁移脚本
│
├── frontend/                Vue 3 前端
│   └── src/
│       ├── api/             fetch 封装 (Bearer + 401)
│       ├── store/           Pinia 认证 store
│       ├── router/          Vue Router + 路由守卫
│       ├── layouts/         侧边栏主布局
│       └── views/           登录/仪表盘/资产/指标/告警
│
├── collector/               分布式采集 Agent
│   └── cmd/agent/           采集器入口
│
├── deploy/
│   ├── docker/              Dockerfile (backend/frontend/collector)
│   ├── compose/             Docker Compose 编排
│   └── k8s/                 Kubernetes (Kustomize + Helm)
│
├── docs/api/                OpenAPI 3.0.3 规范
├── .github/workflows/       CI/CD (构建/测试/镜像推送)
└── .env.example             环境变量模板
```

---

## API 概览

所有 `/api/v1/` 路由需要 `Authorization: Bearer <token>` 头（登录接口除外）。

```
POST   /api/v1/auth/login              登录
GET    /api/v1/auth/me                 当前用户

GET    /api/v1/assets                  资产列表
POST   /api/v1/assets                  创建资产
GET    /api/v1/assets/{id}             资产详情
PUT    /api/v1/assets/{id}             更新资产
DELETE /api/v1/assets/{id}             删除资产
GET    /api/v1/asset-types             资产类型

GET    /api/v1/metrics                 指标定义
POST   /api/v1/collect                 指标上报
GET    /api/v1/metric-data             指标数据查询

GET    /api/v1/alert-rules             告警规则
GET    /api/v1/alert-events            告警事件
POST   /api/v1/alert-events/{id}/ack   确认告警

GET    /api/v1/notify-channels         通知渠道
GET    /api/v1/notify-records          通知记录

GET    /healthz                        健康检查
GET    /readyz                         就绪检查
```

完整 OpenAPI 文档见 [`docs/api/openapi.yaml`](docs/api/openapi.yaml)。

---

## 配置说明

所有配置通过环境变量注入，遵循 12-Factor 原则。

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `APP_ENV` | 运行环境 | development |
| `APP_PORT` | HTTP 端口 | 8080 |
| `APP_LOG_LEVEL` | 日志级别 (debug/info/warn/error) | info |
| `DB_TYPE` | 数据库驱动 (sqlite3/mysql) | sqlite3 |
| `DB_HOST` | 数据库地址 | 127.0.0.1 |
| `DB_PORT` | 数据库端口 | 3306 |
| `DB_NAME` | 数据库名 / SQLite 文件路径 | itsm_ops.db |
| `DB_USER` | 数据库用户 | itsm |
| `DB_PASSWORD` | 数据库密码 | |
| `JWT_SECRET` | JWT 签名密钥 | please-change-me |
| `JWT_EXPIRE` | JWT 过期时间 | 24h |
| `COLLECTOR_ENDPOINT` | 采集器上报地址 | http://127.0.0.1:8080/api/v1/collect |
| `COLLECTOR_TOKEN` | 采集器认证 Token | |
| `COLLECTOR_ASSET_ID` | 采集器关联资产 ID | 1 |
| `COLLECTOR_INTERVAL` | 采集间隔 (秒) | 30 |
| `COLLECTOR_MODE` | 采集模式 (auto/mock) | auto |

完整示例见 [`.env.example`](.env.example)。

---

## 贡献指南

1. Fork 仓库，创建特性分支：`git checkout -b feature/your-feature`
2. 遵循 Conventional Commits 提交规范
3. 确保通过 `go test ./...` 和 `npm run typecheck`
4. 提交 Pull Request

---

## 许可证

[Apache License 2.0](LICENSE)
