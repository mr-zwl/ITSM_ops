# 前端 Web UI (frontend)

基于 Vue 3 + TypeScript + Vite 的运维监控 Web 控制台。

## 目录说明

```
frontend/
├── public/             # 静态资源
├── src/
│   ├── api/            # 后端 API 调用封装
│   ├── assets/         # 图片、字体等
│   ├── components/     # 通用组件
│   ├── layouts/        # 页面布局
│   ├── views/          # 页面视图
│   │   ├── asset/      # 资产管理
│   │   ├── metric/     # 指标管理
│   │   ├── alert/      # 告警管理
│   │   ├── notify/     # 通知管理
│   │   ├── topology/   # 拓扑视图
│   │   ├── report/     # 报表中心
│   │   ├── dashboard/  # 大屏 / 仪表盘
│   │   └── system/     # 系统管理
│   ├── router/         # 路由配置
│   ├── store/          # 状态管理 (Pinia)
│   ├── styles/         # 全局样式
│   ├── hooks/          # 组合式函数
│   └── utils/          # 工具函数
└── tests/              # 单元 / E2E 测试
```
