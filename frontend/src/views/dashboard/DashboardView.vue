<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { get } from '@/api/http'

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

interface AssetMetric {
  asset_id: number
  asset_name: string
  asset_ip: string
  cpu_usage: number | null
  mem_usage: number | null
  disk_usage: number | null
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
const loading = ref(true)
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

const stats = ref<StatCard[]>([])

const assetMetrics = ref<AssetMetric[]>([])

const alerts = ref<AlertItem[]>([])

const alertLevelLabel: Record<string, string> = {
  critical: '严重',
  warning: '警告',
  info: '通知',
}

function severityToLevel(severity: string): 'critical' | 'warning' | 'info' {
  if (severity === 'critical') return 'critical'
  if (severity === 'warning') return 'warning'
  return 'info'
}

function timeAgo(firedAt: string): string {
  const now = new Date()
  const fired = new Date(firedAt)
  const diff = Math.floor((now.getTime() - fired.getTime()) / 1000)
  if (diff < 60) return `${diff} 秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`
  return `${Math.floor(diff / 86400)} 天前`
}

async function fetchDashboard() {
  try {
    const res = await get('/dashboard')
    if (res.code !== 0) return

    const data = res.data as { stats: { total_assets: number; online_assets: number; fired_alerts: number; critical_alerts: number; warning_alerts: number }; metrics: any[]; alerts: any[] }

    // Stats
    stats.value = [
      {
        key: 'assets',
        label: '资产总数',
        value: data.stats.total_assets || 0,
        unit: '台',
        icon: '⬡',
        accent: 'var(--accent-cyan)',
        trend: 'stable',
        trendValue: `${data.stats.online_assets || 0} 在线`,
      },
      {
        key: 'online',
        label: '在线资产',
        value: data.stats.online_assets || 0,
        unit: '台',
        icon: '◈',
        accent: 'var(--accent-teal)',
        trend: (data.stats.online_assets || 0) === (data.stats.total_assets || 0) ? 'stable' : 'down',
        trendValue: (data.stats.online_assets || 0) === (data.stats.total_assets || 0) ? '全部在线' : '部分离线',
      },
      {
        key: 'alerts',
        label: '活跃告警',
        value: data.stats.fired_alerts || 0,
        unit: '条',
        icon: '▲',
        accent: data.stats.critical_alerts > 0 ? 'var(--accent-red)' : 'var(--accent-amber)',
        trend: data.stats.critical_alerts > 0 ? 'up' : 'stable',
        trendValue: data.stats.critical_alerts > 0 ? `${data.stats.critical_alerts} 严重` : '无严重告警',
      },
      {
        key: 'warnings',
        label: '警告',
        value: data.stats.warning_alerts || 0,
        unit: '条',
        icon: '◇',
        accent: 'var(--accent-amber)',
        trend: data.stats.warning_alerts > 0 ? 'up' : 'stable',
        trendValue: data.stats.warning_alerts > 0 ? '需关注' : '正常',
      },
    ]

    // Asset metrics
    assetMetrics.value = (data.metrics || []).map((m: any) => ({
      asset_id: m.asset_id,
      asset_name: m.asset_name,
      asset_ip: m.asset_ip,
      cpu_usage: m.cpu_usage,
      mem_usage: m.mem_usage,
      disk_usage: m.disk_usage,
    }))

    // Alerts
    alerts.value = (data.alerts || []).map((a: any) => ({
      id: a.id,
      level: severityToLevel(a.severity),
      message: a.message,
      source: `资产 #${a.asset_id || '?'}`,
      time: a.fired_at ? timeAgo(a.fired_at) : '-',
    }))

    loading.value = false
  } catch (e) {
    console.error('Failed to fetch dashboard', e)
    loading.value = false
  }
}

onMounted(() => {
  currentTime.value = formatTime()
  timer = setInterval(() => {
    currentTime.value = formatTime()
    uptimeSeconds.value += 1
  }, 1000)

  fetchDashboard()
  // Refresh every 30s
  setInterval(fetchDashboard, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function usageColor(val: number | null): string {
  if (val === null) return 'var(--color-text-muted)'
  if (val >= 95) return 'var(--accent-red)'
  if (val >= 80) return 'var(--accent-amber)'
  return 'var(--accent-emerald)'
}

function usageBarColor(val: number | null): string {
  if (val === null) return 'var(--color-text-muted)'
  if (val >= 95) return 'var(--accent-red)'
  if (val >= 80) return 'var(--accent-amber)'
  if (val >= 60) return 'var(--accent-amber)'
  return 'var(--accent-emerald)'
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
          <div v-if="assetMetrics.length === 0" class="empty-state">
            暂无资产监控数据
          </div>
          <div v-else class="health-sections">
            <div v-for="am in assetMetrics" :key="am.asset_id" class="health-asset">
              <div class="health-asset-header">
                <span class="health-asset-name">{{ am.asset_name }}</span>
                <span class="health-asset-ip">{{ am.asset_ip }}</span>
              </div>
              <div class="health-bars">
                <div class="health-row">
                  <span class="health-label">CPU</span>
                  <div class="health-track">
                    <div class="health-fill" :style="{ width: (am.cpu_usage || 0) + '%', background: usageBarColor(am.cpu_usage) }"></div>
                  </div>
                  <span class="health-value" :style="{ color: usageColor(am.cpu_usage) }">{{ am.cpu_usage !== null ? am.cpu_usage.toFixed(1) + '%' : '-' }}</span>
                </div>
                <div class="health-row">
                  <span class="health-label">内存</span>
                  <div class="health-track">
                    <div class="health-fill" :style="{ width: (am.mem_usage || 0) + '%', background: usageBarColor(am.mem_usage) }"></div>
                  </div>
                  <span class="health-value" :style="{ color: usageColor(am.mem_usage) }">{{ am.mem_usage !== null ? am.mem_usage.toFixed(1) + '%' : '-' }}</span>
                </div>
                <div class="health-row">
                  <span class="health-label">磁盘</span>
                  <div class="health-track">
                    <div class="health-fill" :style="{ width: (am.disk_usage || 0) + '%', background: usageBarColor(am.disk_usage) }"></div>
                  </div>
                  <span class="health-value" :style="{ color: usageColor(am.disk_usage) }">{{ am.disk_usage !== null ? am.disk_usage.toFixed(1) + '%' : '-' }}</span>
                </div>
              </div>
            </div>
          </div>
        </article>

        <article class="detail-card detail-card--alerts">
          <h2 class="detail-card__title">
            <span class="title-icon" style="color: var(--accent-red)">▲</span>
            最新告警
          </h2>
          <div v-if="alerts.length === 0" class="empty-state">
            ✅ 暂无活跃告警，系统运行正常
          </div>
          <ul v-else class="alert-list">
            <li
              v-for="alert in alerts"
              :key="alert.id"
              class="alert-item"
              :class="`alert-item--${alert.level}`"
            >
              <span class="alert-level" :class="`alert-level--${alert.level}`">
                {{ alertLevelLabel[alert.level] || alert.level }}
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
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.1);
}

.trend--down {
  color: var(--accent-amber);
  background: rgba(240, 160, 48, 0.1);
}

.trend--stable {
  color: var(--accent-emerald);
  background: rgba(45, 212, 160, 0.1);
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
.health-sections {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.health-asset {
  padding: var(--space-4);
  background: var(--color-bg-deep);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-subtle);
}

.health-asset-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-3);
}

.health-asset-name {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.health-asset-ip {
  font-family: var(--font-display);
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.health-bars {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.health-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.health-label {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  width: 36px;
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
  transition: width 0.6s ease;
}

.health-value {
  font-family: var(--font-display);
  font-size: 0.8rem;
  width: 52px;
  text-align: right;
  flex-shrink: 0;
}

.empty-state {
  text-align: center;
  padding: var(--space-8) 0;
  color: var(--color-text-muted);
  font-size: 0.85rem;
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
    width: 36px;
    font-size: 0.75rem;
  }

  .alert-message {
    font-size: 0.75rem;
  }
}
</style>
