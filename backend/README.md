# 后端服务 (backend)

基于 Go 的 RESTful API 服务，提供资产管理、指标管理、告警引擎、通知中心、拓扑、报表、系统管理等核心能力。

## 目录说明

```
backend/
├── cmd/                # 程序入口 (main.go)
├── internal/           # 内部业务代码（不对外暴露）
│   ├── api/            # HTTP 路由 & Handler
│   ├── asset/          # 资产管理模块
│   ├── metric/         # 指标管理模块
│   ├── alert/          # 告警引擎模块
│   ├── notify/         # 通知中心模块
│   ├── topology/       # 拓扑模块
│   ├── report/         # 报表模块
│   ├── auth/           # 认证 & 鉴权
│   ├── config/         # 配置加载
│   ├── middleware/     # HTTP 中间件
│   ├── model/          # 数据模型 (DTO / Entity)
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
├── pkg/                # 可复用组件
│   ├── logger/         # 日志
│   ├── db/             # 关系数据库 (MySQL/PG)
│   ├── tsdb/           # 时序数据库 (InfluxDB/VictoriaMetrics)
│   ├── mq/             # 消息队列
│   ├── cache/          # 缓存 (Redis)
│   └── utils/          # 工具函数
├── configs/            # 配置文件示例
├── migrations/         # 数据库迁移 SQL
└── tests/              # 集成测试
```
