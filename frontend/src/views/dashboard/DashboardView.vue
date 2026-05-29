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
// uptimeSeconds removed
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

    const data = res.data as { stats: { total_assets: number; online_assets: number; fired_alerts: number; critical_alerts: number; warning_alerts: number; total_metrics: number }; metrics: any[]; recent_alerts: any[] }

    stats.value = [
      {
        key: 'assets',
        label: '资产总数',
        value: data.stats.total_assets || 0,
        unit: '台',
        icon: '💻',
        accent: 'blue',
        trend: 'stable',
        trendValue: `${data.stats.online_assets || 0} 在线`,
      },
      {
        key: 'online',
        label: '在线资产',
        value: data.stats.online_assets || 0,
        unit: '台',
        icon: '✅',
        accent: 'green',
        trend: (data.stats.online_assets || 0) === (data.stats.total_assets || 0) ? 'stable' : 'down',
        trendValue: (data.stats.online_assets || 0) === (data.stats.total_assets || 0) ? '全部在线' : '部分离线',
      },
      {
        key: 'alerts',
        label: '活跃告警',
        value: data.stats.fired_alerts || 0,
        unit: '条',
        icon: '🔔',
        accent: data.stats.critical_alerts > 0 ? 'red' : 'amber',
        trend: data.stats.critical_alerts > 0 ? 'up' : 'stable',
        trendValue: data.stats.critical_alerts > 0 ? `${data.stats.critical_alerts} 严重` : '无严重告警',
      },
      {
        key: 'metrics',
        label: '监控指标',
        value: data.stats.total_metrics || 0,
        unit: '条',
        icon: '📈',
        accent: 'blue',
        trend: 'stable',
        trendValue: '持续采集中',
      },
    ]

    assetMetrics.value = (data.metrics || []).map((m: any) => ({
      asset_id: m.asset_id,
      asset_name: m.asset_name,
      asset_ip: m.asset_ip,
      cpu_usage: m.cpu_usage,
      mem_usage: m.mem_usage,
      disk_usage: m.disk_usage,
    }))

    alerts.value = (data.recent_alerts || []).map((a: any) => ({
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
    // uptimeSeconds removed
  }, 1000)
  fetchDashboard()
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

function usageBarStyle(val: number | null): string {
  if (val === null) return '0%'
  return `${Math.min(val, 100)}%`
}

function statCardClass(accent: string): string {
  return `stat-card stat-card--${accent}`
}
</script>

<template>
  <div class="dashboard">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">仪表盘</h1>
      <p class="page-desc">系统运行概览</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div v-for="s in stats" :key="s.key" :class="statCardClass(s.accent)">
        <div class="stat-icon">{{ s.icon }}</div>
        <div class="stat-body">
          <span class="stat-label">{{ s.label }}</span>
          <div class="stat-value-row">
            <span class="stat-value">{{ s.value }}</span>
            <span class="stat-unit">{{ s.unit }}</span>
          </div>
          <span class="stat-trend" :class="`trend--${s.trend}`">{{ s.trendValue }}</span>
        </div>
      </div>
    </div>

    <!-- 双栏内容 -->
    <div class="content-grid">
      <!-- 资产健康 -->
      <div class="content-card">
        <div class="card-header">
          <h2 class="card-title">🖥️ 资产健康</h2>
        </div>
        <div class="health-list">
          <div v-for="m in assetMetrics" :key="m.asset_id" class="health-item">
            <div class="health-info">
              <span class="health-name">{{ m.asset_name }}</span>
              <span class="health-ip">{{ m.asset_ip }}</span>
            </div>
            <div class="health-bars">
              <div class="bar-row">
                <span class="bar-label">CPU</span>
                <div class="bar-track">
                  <div class="bar-fill bar-fill--cpu" :style="{ width: usageBarStyle(m.cpu_usage) }"></div>
                </div>
                <span class="bar-value" :style="{ color: usageColor(m.cpu_usage) }">{{ m.cpu_usage !== null ? m.cpu_usage.toFixed(1) + '%' : '-' }}</span>
              </div>
              <div class="bar-row">
                <span class="bar-label">内存</span>
                <div class="bar-track">
                  <div class="bar-fill bar-fill--mem" :style="{ width: usageBarStyle(m.mem_usage) }"></div>
                </div>
                <span class="bar-value" :style="{ color: usageColor(m.mem_usage) }">{{ m.mem_usage !== null ? m.mem_usage.toFixed(1) + '%' : '-' }}</span>
              </div>
              <div class="bar-row">
                <span class="bar-label">磁盘</span>
                <div class="bar-track">
                  <div class="bar-fill bar-fill--disk" :style="{ width: usageBarStyle(m.disk_usage) }"></div>
                </div>
                <span class="bar-value" :style="{ color: usageColor(m.disk_usage) }">{{ m.disk_usage !== null ? m.disk_usage.toFixed(1) + '%' : '-' }}</span>
              </div>
            </div>
          </div>
          <div v-if="assetMetrics.length === 0" class="empty-hint">暂无资产数据</div>
        </div>
      </div>

      <!-- 最近告警 -->
      <div class="content-card">
        <div class="card-header">
          <h2 class="card-title">🔔 最近告警</h2>
        </div>
        <div class="alert-list">
          <div v-for="a in alerts" :key="a.id" class="alert-item" :class="`alert-item--${a.level}`">
            <span class="alert-level" :class="`xhs-tag xhs-tag--${a.level === 'critical' ? 'red' : a.level === 'warning' ? 'amber' : 'blue'}`">{{ alertLevelLabel[a.level] }}</span>
            <div class="alert-content">
              <span class="alert-message">{{ a.message }}</span>
              <span class="alert-meta">{{ a.source }} · {{ a.time }}</span>
            </div>
          </div>
          <div v-if="alerts.length === 0" class="empty-hint">🎉 暂无告警，一切正常</div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.page-header {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.page-desc {
  font-size: 0.9rem;
  color: var(--color-text-muted);
}

/* ====== Stats Grid ====== */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
}

.stat-card {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  transition: all var(--transition-fast);
  position: relative;
  overflow: hidden;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-card-hover);
}

.stat-card--blue { border-left: 3px solid var(--accent-cyan); }
.stat-card--green { border-left: 3px solid var(--accent-emerald); }
.stat-card--red { border-left: 3px solid var(--accent-red); }
.stat-card--amber { border-left: 3px solid var(--accent-amber); }

.stat-icon {
  font-size: 2rem;
  line-height: 1;
  flex-shrink: 0;
}

.stat-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.stat-label {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  font-weight: 500;
}

.stat-value-row {
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1.1;
  color: var(--color-text-primary);
}

.stat-unit {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.stat-trend {
  font-size: 0.75rem;
  font-weight: 500;
}

.trend--up { color: var(--accent-red); }
.trend--down { color: var(--accent-amber); }
.trend--stable { color: var(--color-text-muted); }

/* ====== Content Grid ====== */
.content-grid {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: var(--space-4);
}

.content-card {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

.card-header {
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.card-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* ====== Health List ====== */
.health-list {
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  max-height: 500px;
  overflow-y: auto;
}

.health-item {
  padding: var(--space-4);
  background: var(--color-bg-elevated);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.health-info {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.health-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.health-ip {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.health-bars {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.bar-row {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.bar-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  width: 32px;
  flex-shrink: 0;
}

.bar-track {
  flex: 1;
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width var(--transition-base);
}

.bar-fill--cpu { background: var(--accent-cyan); }
.bar-fill--mem { background: var(--accent-emerald); }
.bar-fill--disk { background: var(--accent-amber); }

.bar-value {
  font-size: 0.75rem;
  font-weight: 600;
  width: 44px;
  text-align: right;
  flex-shrink: 0;
}

/* ====== Alert List ====== */
.alert-list {
  padding: var(--space-4);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  max-height: 500px;
  overflow-y: auto;
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--color-bg-elevated);
}

.alert-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.alert-message {
  font-size: 0.85rem;
  color: var(--color-text-primary);
  line-height: 1.5;
}

.alert-meta {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.empty-hint {
  text-align: center;
  padding: var(--space-8) var(--space-4);
  color: var(--color-text-muted);
  font-size: 0.9rem;
}

@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  .content-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
