<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get, post } from '@/api/http'

type Severity = 'critical' | 'warning' | 'info'
type AlertStatus = 'firing' | 'acked'

interface AlertEvent {
  id: number
  rule_id: number
  asset_id: number
  severity: Severity
  message: string
  current_val: number
  status: AlertStatus
  fired_at: string
}

const loading = ref(true)
const error = ref('')
const alerts = ref<AlertEvent[]>([])
const ackingIds = ref(new Set<number>())
const filterStatus = ref<string>('')

const severityLabel: Record<Severity, string> = {
  critical: '严重',
  warning: '警告',
  info: '通知',
}

const statusLabel: Record<AlertStatus, string> = {
  firing: '告警中',
  acked: '已确认',
}

function severityColor(sev: Severity): string {
  if (sev === 'critical') return 'red'
  if (sev === 'warning') return 'amber'
  return 'blue'
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

async function fetchAlerts(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const url = filterStatus.value ? `/alert-events?status=${filterStatus.value}` : '/alert-events'
    const res = await get<AlertEvent[]>(url)
    alerts.value = res.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载告警事件失败'
  } finally {
    loading.value = false
  }
}

async function ackAlert(id: number): Promise<void> {
  ackingIds.value.add(id)
  try {
    await post<{ status: string }>(`/alert-events/${id}/ack`, {})
    const target = alerts.value.find((a) => a.id === id)
    if (target) {
      target.status = 'acked'
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '确认告警失败'
  } finally {
    ackingIds.value.delete(id)
  }
}

onMounted(fetchAlerts)
</script>

<template>
  <div class="alert-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">告警事件</h1>
        <p class="page-desc">查看和处理系统告警</p>
      </div>
      <div class="filter-bar">
        <button class="filter-btn" :class="{ 'filter-btn--active': filterStatus === '' }" @click="filterStatus = ''; fetchAlerts()">全部</button>
        <button class="filter-btn" :class="{ 'filter-btn--active': filterStatus === 'firing' }" @click="filterStatus = 'firing'; fetchAlerts()">告警中</button>
        <button class="filter-btn" :class="{ 'filter-btn--active': filterStatus === 'acked' }" @click="filterStatus = 'acked'; fetchAlerts()">已确认</button>
      </div>
    </div>

    <div v-if="error" class="form-error">{{ error }}</div>
    <div v-if="loading" class="loading-hint">加载中...</div>

    <div v-else class="alert-list">
      <div v-for="alert in alerts" :key="alert.id" class="alert-card" :class="`alert-card--${alert.severity}`">
        <div class="alert-card-top">
          <span class="xhs-tag" :class="`xhs-tag--${severityColor(alert.severity)}`">{{ severityLabel[alert.severity] }}</span>
          <span class="alert-status" :class="`alert-status--${alert.status}`">{{ statusLabel[alert.status] }}</span>
          <span class="alert-time">{{ formatTime(alert.fired_at) }}</span>
        </div>

        <div class="alert-card-body">
          <p class="alert-message">{{ alert.message }}</p>
          <div class="alert-meta">
            <span>规则 #{{ alert.rule_id }}</span>
            <span>资产 #{{ alert.asset_id || '?' }}</span>
            <span>当前值: {{ alert.current_val.toFixed(2) }}</span>
          </div>
        </div>

        <div class="alert-card-actions" v-if="alert.status === 'firing'">
          <button class="btn-ack" :disabled="ackingIds.has(alert.id)" @click="ackAlert(alert.id)">
            {{ ackingIds.has(alert.id) ? '处理中...' : '✓ 确认告警' }}
          </button>
        </div>
      </div>

      <div v-if="alerts.length === 0" class="empty-state">
        <span class="empty-icon">🎉</span>
        <p>暂无告警，一切正常</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.alert-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.page-desc {
  font-size: 0.9rem;
  color: var(--color-text-muted);
  margin-top: 2px;
}

.filter-bar {
  display: flex;
  gap: var(--space-2);
}

.filter-btn {
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border);
  border-radius: 100px;
  background: var(--color-bg-base);
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
}

.filter-btn:hover {
  border-color: var(--accent-red);
  color: var(--accent-red);
}

.filter-btn--active {
  background: var(--accent-red);
  border-color: var(--accent-red);
  color: white;
}

.form-error {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--accent-pink);
  color: var(--accent-red);
  font-size: 0.85rem;
  font-weight: 500;
}

/* ====== Alert Cards ====== */
.alert-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.alert-card {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.alert-card:hover {
  box-shadow: var(--shadow-card-hover);
}

.alert-card--critical {
  border-left: 4px solid var(--accent-red);
}

.alert-card--warning {
  border-left: 4px solid var(--accent-amber);
}

.alert-card--info {
  border-left: 4px solid var(--accent-cyan);
}

.alert-card-top {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-5);
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg-elevated);
}

.alert-status {
  font-size: 0.75rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 100px;
}

.alert-status--firing {
  color: var(--accent-red);
  background: var(--accent-pink);
}

.alert-status--acked {
  color: var(--accent-emerald);
  background: var(--accent-emerald-light);
}

.alert-time {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  margin-left: auto;
}

.alert-card-body {
  padding: var(--space-4) var(--space-5);
}

.alert-message {
  font-size: 0.9rem;
  color: var(--color-text-primary);
  line-height: 1.6;
  font-weight: 500;
}

.alert-meta {
  display: flex;
  gap: var(--space-4);
  margin-top: var(--space-2);
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.alert-card-actions {
  padding: var(--space-3) var(--space-5);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-elevated);
}

.btn-ack {
  padding: var(--space-2) var(--space-5);
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-emerald);
  color: white;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
}

.btn-ack:hover:not(:disabled) {
  background: #00B347;
}

.btn-ack:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.empty-state {
  text-align: center;
  padding: var(--space-12) var(--space-4);
  color: var(--color-text-muted);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: var(--space-3);
}

.loading-hint {
  text-align: center;
  padding: var(--space-10);
  color: var(--color-text-muted);
}
</style>
