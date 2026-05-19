# 分布式采集 Agent (collector)

独立部署的采集器，负责通过 SNMP / SSH / IPMI / WMI / Agent / API 等协议采集设备指标，上报至后端。

## 目录说明

```
collector/
├── cmd/                # 入口
├── internal/
│   ├── snmp/           # SNMP 协议采集
│   ├── ssh/            # SSH 命令采集
│   ├── ipmi/           # IPMI 带外采集
│   ├── wmi/            # Windows WMI 采集
│   ├── agent/          # Agent 模式（设备内安装）
│   ├── api/            # 厂商 API 对接
│   ├── scheduler/      # 任务调度
│   └── reporter/       # 数据上报
├── pkg/
│   ├── protocol/       # 协议抽象
│   ├── template/       # 采集模板引擎
│   └── utils/
├── configs/            # 配置示例
├── templates/          # 厂商采集模板
│   ├── cisco/
│   ├── huawei/
│   ├── h3c/
│   ├── dell/
│   ├── hpe/
│   ├── inspur/
│   ├── linux/
│   └── windows/
└── tests/
```

## 部署模式

- **Sidecar / DaemonSet**（K8s 场景）：每个采集区域部署一个
- **独立进程**（传统场景）：通过 systemd / docker 启动
- **嵌入式**：编译为单二进制，直接在边缘节点运行
