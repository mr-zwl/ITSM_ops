<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

interface StatCard {
  key: string
  label: string
  value: number
  unit: string
  icon: string
  accent: string
  trend: 'up' | 'down' | 'stable'
  trendValue: string
}

interface AlertItem {
  id: number
  level: 'critical' | 'warning' | 'info'
  message: string
  source: string
  time: string
}

const currentTime = ref('')
const uptimeSeconds = ref(0)
let timer: ReturnType<typeof setInterval> | undefined

function formatTime(): string {
  const now = new Date()
  return now.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60
  return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
}

onMounted(() => {
  currentTime.value = formatTime()
  timer = setInterval(() => {
    currentTime.value = formatTime()
    uptimeSeconds.value += 1
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

const stats = ref<StatCard[]>([
  {
    key: 'servers',
    label: '服务器',
    value: 128,
    unit: '台',
    icon: '⬡',
    accent: 'var(--accent-cyan)',
    trend: 'up',
    trendValue: '+3',
  },
  {
    key: 'switches',
    label: '交换机',
    value: 64,
    unit: '台',
    icon: '◈',
    accent: 'var(--accent-teal)',
    trend: 'stable',
    trendValue: '0',
  },
  {
    key: 'storage',
    label: '存储集群',
    value: 12,
    unit: '套',
    icon: '◎',
    accent: 'var(--accent-amber)',
    trend: 'up',
    trendValue: '+1',
  },
  {
    key: 'metrics',
    label: '监控指标',
    value: 2847,
    unit: '项',
    icon: '◇',
    accent: 'var(--accent-emerald)',
    trend: 'up',
    trendValue: '+156',
  },
])

const alerts = ref<AlertItem[]>([
  { id: 1, level: 'critical', message: '生产数据库主节点 DB-Master-01 连接超时', source: 'db-cluster-prod', time: '2 分钟前' },
  { id: 2, level: 'warning', message: '交换机 SW-Core-A3 端口利用率超过 85%', source: 'network-core', time: '8 分钟前' },
  { id: 3, level: 'critical', message: '存储集群 NAS-03 磁盘阵列降级告警', source: 'storage-nas', time: '15 分钟前' },
  { id: 4, level: 'warning', message: '应用服务器 APP-Web-12 内存使用率 92%', source: 'app-cluster', time: '22 分钟前' },
  { id: 5, level: 'info', message: '备份任务 BK-Daily 完成耗时超出预期窗口', source: 'backup-scheduler', time: '45 分钟前' },
  { id: 6, level: 'info', message: 'SSL 证书 itsm.example.com 将于 15 天后过期', source: 'cert-monitor', time: '1 小时前' },
])

const alertLevelLabel: Record<AlertItem['level'], string> = {
  critical: '严重',
  warning: '警告',
  info: '通知',
}
</script>

<template>
  <div class="dashboard">
    <header class="dashboard-header">
      <div class="header-left">
        <div class="status-beacon">
          <span class="beacon-dot"></span>
          <span class="beacon-text">系统运行中</span>
        </div>
        <span class="header-uptime">运行时长 {{ formatUptime(uptimeSeconds) }}</span>
      </div>
      <div class="header-right">
        <time class="header-time">{{ currentTime }}</time>
      </div>
    </header>

    <main class="dashboard-body">
      <section class="stats-grid">
        <article
          v-for="stat in stats"
          :key="stat.key"
          class="stat-card"
          :class="`stat-card--${stat.key}`"
        >
          <div class="stat-card__icon" :style="{ color: stat.accent }">{{ stat.icon }}</div>
          <div class="stat-card__body">
            <span class="stat-card__label">{{ stat.label }}</span>
            <div class="stat-card__value-row">
              <span class="stat-card__value">{{ stat.value.toLocaleString() }}</span>
              <span class="stat-card__unit">{{ stat.unit }}</span>
            </div>
          </div>
          <div class="stat-card__trend" :class="`trend--${stat.trend}`">
            {{ stat.trendValue }}
          </div>
          <div class="stat-card__glow" :style="{ background: stat.accent }"></div>
        </article>
      </section>

      <section class="detail-grid">
        <article class="detail-card detail-card--health">
          <h2 class="detail-card__title">
            <span class="title-icon" style="color: var(--accent-emerald)">◉</span>
            资源健康概览
          </h2>
          <div class="health-bars">
            <div class="health-row">
              <span class="health-label">CPU 使用率</span>
              <div class="health-track">
                <div class="health-fill health-fill--cpu" style="width: 67%"></div>
              </div>
              <span class="health-value">67%</span>
            </div>
            <div class="health-row">
              <span class="health-label">内存使用率</span>
              <div class="health-track">
                <div class="health-fill health-fill--mem" style="width: 74%"></div>
              </div>
              <span class="health-value">74%</span>
            </div>
            <div class="health-row">
              <span class="health-label">磁盘 I/O</span>
              <div class="health-track">
                <div class="health-fill health-fill--disk" style="width: 42%"></div>
              </div>
              <span class="health-value">42%</span>
            </div>
            <div class="health-row">
              <span class="health-label">网络带宽</span>
              <div class="health-track">
                <div class="health-fill health-fill--net" style="width: 58%"></div>
              </div>
              <span class="health-value">58%</span>
            </div>
          </div>
        </article>

        <article class="detail-card detail-card--alerts">
          <h2 class="detail-card__title">
            <span class="title-icon" style="color: var(--accent-red)">▲</span>
            最新告警
          </h2>
          <ul class="alert-list">
            <li
              v-for="alert in alerts"
              :key="alert.id"
              class="alert-item"
              :class="`alert-item--${alert.level}`"
            >
              <span class="alert-level" :class="`alert-level--${alert.level}`">
                {{ alertLevelLabel[alert.level] }}
              </span>
              <div class="alert-content">
                <span class="alert-message">{{ alert.message }}</span>
                <span class="alert-meta">{{ alert.source }} · {{ alert.time }}</span>
              </div>
            </li>
          </ul>
        </article>
      </section>

      <footer class="dashboard-footer">
        <span>ITSM Ops v0.1.0</span>
        <span class="footer-sep">|</span>
        <span>智能运维监控平台</span>
      </footer>
    </main>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-5);
}

.status-beacon {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.beacon-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-emerald);
  box-shadow: 0 0 6px var(--accent-emerald), 0 0 16px rgba(45, 212, 160, 0.3);
  animation: beacon-pulse 2s ease-in-out infinite;
}

@keyframes beacon-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.beacon-text {
  font-size: 0.8rem;
  color: var(--accent-emerald);
  font-weight: 500;
}

.header-uptime {
  font-family: var(--font-display);
  font-size: 0.75rem;
  color: var(--color-text-muted);
  letter-spacing: 0.04em;
}

.header-right {
  text-align: right;
}

.header-time {
  font-family: var(--font-display);
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  letter-spacing: 0.06em;
}

/* ====== Stats Grid ====== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-5);
}

.stat-card {
  position: relative;
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5) var(--space-6);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  transition: transform var(--transition-fast), border-color var(--transition-base);
}

.stat-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-border-subtle);
}

.stat-card__icon {
  font-size: 2rem;
  line-height: 1;
  flex-shrink: 0;
  filter: drop-shadow(0 0 8px currentColor);
}

.stat-card__body {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.stat-card__label {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.stat-card__value-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
}

.stat-card__value {
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1;
  color: var(--color-text-primary);
}

.stat-card__unit {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.stat-card__trend {
  position: absolute;
  top: var(--space-3);
  right: var(--space-4);
  font-family: var(--font-display);
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  letter-spacing: 0.02em;
}

.trend--up {
  color: var(--accent-emerald);
  background: rgba(45, 212, 160, 0.1);
}

.trend--down {
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.1);
}

.trend--stable {
  color: var(--color-text-muted);
  background: rgba(77, 98, 130, 0.15);
}

.stat-card__glow {
  position: absolute;
  bottom: -20px;
  right: -20px;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  opacity: 0.06;
  filter: blur(24px);
  pointer-events: none;
}

/* ====== Detail Grid ====== */
.detail-grid {
  display: grid;
  grid-template-columns: 1fr 1.4fr;
  gap: var(--space-5);
}

.detail-card {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-6);
  box-shadow: var(--shadow-card);
}

.detail-card__title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--space-5);
  letter-spacing: 0.02em;
}

.title-icon {
  font-size: 0.75rem;
}

/* ====== Health Bars ====== */
.health-bars {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.health-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.health-label {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  width: 80px;
  flex-shrink: 0;
  text-align: right;
}

.health-track {
  flex: 1;
  height: 8px;
  background: var(--color-bg-deep);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.health-fill {
  height: 100%;
  border-radius: var(--radius-sm);
  transition: width var(--transition-base);
}

.health-fill--cpu {
  background: linear-gradient(90deg, var(--accent-cyan), rgba(0, 212, 255, 0.6));
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.3);
}

.health-fill--mem {
  background: linear-gradient(90deg, var(--accent-teal), rgba(14, 165, 160, 0.6));
  box-shadow: 0 0 8px rgba(14, 165, 160, 0.3);
}

.health-fill--disk {
  background: linear-gradient(90deg, var(--accent-amber), rgba(240, 160, 48, 0.6));
  box-shadow: 0 0 8px rgba(240, 160, 48, 0.3);
}

.health-fill--net {
  background: linear-gradient(90deg, var(--accent-emerald), rgba(45, 212, 160, 0.6));
  box-shadow: 0 0 8px rgba(45, 212, 160, 0.3);
}

.health-value {
  font-family: var(--font-display);
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  width: 40px;
  text-align: right;
  flex-shrink: 0;
}

/* ====== Alert List ====== */
.alert-list {
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--color-bg-deep);
  border: 1px solid transparent;
  transition: border-color var(--transition-fast);
}

.alert-item:hover {
  border-color: var(--color-border);
}

.alert-item--critical {
  border-left: 3px solid var(--accent-red);
}

.alert-item--warning {
  border-left: 3px solid var(--accent-amber);
}

.alert-item--info {
  border-left: 3px solid var(--accent-cyan);
}

.alert-level {
  flex-shrink: 0;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  text-transform: uppercase;
  line-height: 1.4;
  margin-top: 2px;
}

.alert-level--critical {
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.12);
}

.alert-level--warning {
  color: var(--accent-amber);
  background: rgba(240, 160, 48, 0.12);
}

.alert-level--info {
  color: var(--accent-cyan);
  background: rgba(0, 212, 255, 0.1);
}

.alert-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.alert-message {
  font-size: 0.8rem;
  color: var(--color-text-primary);
  line-height: 1.5;
}

.alert-meta {
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

/* ====== Footer ====== */
.dashboard-footer {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4) 0;
  font-size: 0.7rem;
  color: var(--color-text-muted);
  border-top: 1px solid var(--color-border-subtle);
  margin-top: auto;
}

.footer-sep {
  opacity: 0.3;
}

/* ====== Responsive ====== */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .dashboard-header {
    flex-direction: column;
    gap: var(--space-3);
    padding: var(--space-4);
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .stat-card__value {
    font-size: 1.5rem;
  }

  .health-label {
    width: 64px;
    font-size: 0.75rem;
  }

  .alert-message {
    font-size: 0.75rem;
  }
}
</style>
