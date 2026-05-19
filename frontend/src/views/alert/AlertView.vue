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
  if (sev === 'critical') return 'var(--accent-red)'
  if (sev === 'warning') return 'var(--accent-amber)'
  return 'var(--accent-cyan)'
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
    const res = await get<AlertEvent[]>('/alert-events')
    alerts.value = res.data
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
    <header class="page-header">
      <h1 class="page-title">
        <span class="title-icon" style="color: var(--accent-red)">▲</span>
        告警事件
      </h1>
      <span class="page-subtitle">实时告警与事件处理</span>
    </header>

    <div v-if="loading" class="state-panel">
      <div class="loading-ring"></div>
      <span class="state-text">加载中...</span>
    </div>

    <div v-else-if="error" class="state-panel state-panel--error">
      <span class="state-icon">⚠</span>
      <span class="state-text">{{ error }}</span>
      <button class="retry-btn" @click="fetchAlerts">重试</button>
    </div>

    <div v-else-if="alerts.length === 0" class="state-panel">
      <span class="state-icon">✓</span>
      <span class="state-text">暂无告警事件</span>
    </div>

    <template v-else>
      <div class="alert-summary">
        <div class="summary-chip summary-chip--firing">
          <span class="summary-dot summary-dot--firing"></span>
          告警中 {{ alerts.filter((a) => a.status === 'firing').length }}
        </div>
        <div class="summary-chip summary-chip--acked">
          <span class="summary-dot summary-dot--acked"></span>
          已确认 {{ alerts.filter((a) => a.status === 'acked').length }}
        </div>
        <div class="summary-chip summary-chip--total">
          共计 {{ alerts.length }}
        </div>
      </div>

      <div class="alert-table-wrap">
        <table class="alert-table">
          <thead>
            <tr>
              <th class="col-severity">级别</th>
              <th class="col-message">告警信息</th>
              <th class="col-value">当前值</th>
              <th class="col-status">状态</th>
              <th class="col-time">触发时间</th>
              <th class="col-action">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="alert in alerts"
              :key="alert.id"
              class="alert-row"
              :class="[
                `alert-row--${alert.severity}`,
                { 'alert-row--acked': alert.status === 'acked' },
              ]"
            >
              <td class="col-severity">
                <span
                  class="severity-badge"
                  :class="`severity-badge--${alert.severity}`"
                >
                  {{ severityLabel[alert.severity] }}
                </span>
              </td>
              <td class="col-message">
                <span class="alert-msg">{{ alert.message }}</span>
                <span class="alert-meta-inline">资产 #{{ alert.asset_id }}</span>
              </td>
              <td class="col-value">
                <span class="val-text" :style="{ color: severityColor(alert.severity) }">
                  {{ alert.current_val?.toFixed(1) ?? '-' }}
                </span>
              </td>
              <td class="col-status">
                <span
                  class="status-tag"
                  :class="`status-tag--${alert.status}`"
                >
                  <span class="status-dot" :class="`status-dot--${alert.status}`"></span>
                  {{ statusLabel[alert.status] }}
                </span>
              </td>
              <td class="col-time">
                <span class="time-text">{{ formatTime(alert.fired_at) }}</span>
              </td>
              <td class="col-action">
                <button
                  v-if="alert.status === 'firing'"
                  class="ack-btn"
                  :disabled="ackingIds.has(alert.id)"
                  @click="ackAlert(alert.id)"
                >
                  {{ ackingIds.has(alert.id) ? '确认中...' : '确认' }}
                </button>
                <span v-else class="acked-label">已处理</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.alert-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* ====== Page Header ====== */
.page-header {
  display: flex;
  align-items: baseline;
  gap: var(--space-4);
}

.page-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: 0.02em;
}

.title-icon {
  font-size: 0.9rem;
}

.page-subtitle {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

/* ====== State Panels ====== */
.state-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-4);
  padding: var(--space-12) var(--space-6);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
}

.state-panel--error {
  border-color: rgba(240, 72, 72, 0.3);
}

.state-icon {
  font-size: 2rem;
  filter: drop-shadow(0 0 8px currentColor);
}

.state-text {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
}

.loading-ring {
  width: 32px;
  height: 32px;
  border: 2px solid var(--color-border);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.retry-btn {
  padding: var(--space-2) var(--space-5);
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid rgba(0, 212, 255, 0.3);
  border-radius: var(--radius-md);
  color: var(--accent-cyan);
  font-size: 0.8rem;
  font-family: var(--font-body);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.retry-btn:hover {
  background: rgba(0, 212, 255, 0.18);
  border-color: var(--accent-cyan);
}

/* ====== Alert Summary ====== */
.alert-summary {
  display: flex;
  gap: var(--space-3);
  flex-wrap: wrap;
}

.summary-chip {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 0.78rem;
  font-weight: 500;
  padding: var(--space-2) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
}

.summary-chip--firing {
  color: var(--accent-red);
  border-color: rgba(240, 72, 72, 0.2);
}

.summary-chip--acked {
  color: var(--accent-emerald);
  border-color: rgba(45, 212, 160, 0.2);
}

.summary-chip--total {
  color: var(--color-text-secondary);
}

.summary-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.summary-dot--firing {
  background: var(--accent-red);
  box-shadow: 0 0 6px rgba(240, 72, 72, 0.5);
  animation: blink 1.5s ease-in-out infinite;
}

.summary-dot--acked {
  background: var(--accent-emerald);
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}

/* ====== Alert Table ====== */
.alert-table-wrap {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
}

.alert-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.82rem;
}

.alert-table thead {
  background: var(--color-bg-elevated);
}

.alert-table th {
  padding: var(--space-3) var(--space-4);
  text-align: left;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  border-bottom: 1px solid var(--color-border);
  white-space: nowrap;
}

.alert-table td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border-subtle);
  vertical-align: middle;
}

.alert-row {
  transition: background var(--transition-fast);
}

.alert-row:hover {
  background: rgba(0, 212, 255, 0.03);
}

.alert-row:last-child td {
  border-bottom: none;
}

.alert-row--critical {
  border-left: 3px solid var(--accent-red);
}

.alert-row--warning {
  border-left: 3px solid var(--accent-amber);
}

.alert-row--info {
  border-left: 3px solid var(--accent-cyan);
}

.alert-row--acked {
  opacity: 0.55;
}

/* ====== Severity Badge ====== */
.severity-badge {
  display: inline-block;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.06em;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  line-height: 1.5;
}

.severity-badge--critical {
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.12);
}

.severity-badge--warning {
  color: var(--accent-amber);
  background: rgba(240, 160, 48, 0.12);
}

.severity-badge--info {
  color: var(--accent-cyan);
  background: rgba(0, 212, 255, 0.1);
}

/* ====== Message Column ====== */
.col-message {
  min-width: 200px;
}

.alert-msg {
  display: block;
  color: var(--color-text-primary);
  line-height: 1.5;
}

.alert-meta-inline {
  display: block;
  font-size: 0.68rem;
  color: var(--color-text-muted);
  margin-top: 2px;
}

/* ====== Value Column ====== */
.val-text {
  font-family: var(--font-display);
  font-size: 0.82rem;
  font-weight: 600;
}

/* ====== Status Tag ====== */
.status-tag {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot--firing {
  background: var(--accent-red);
  box-shadow: 0 0 6px rgba(240, 72, 72, 0.5);
  animation: blink 1.5s ease-in-out infinite;
}

.status-dot--acked {
  background: var(--accent-emerald);
  box-shadow: 0 0 4px rgba(45, 212, 160, 0.3);
}

.status-tag--firing {
  color: var(--accent-red);
}

.status-tag--acked {
  color: var(--accent-emerald);
}

/* ====== Time Column ====== */
.time-text {
  font-family: var(--font-display);
  font-size: 0.75rem;
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
  white-space: nowrap;
}

/* ====== Ack Button ====== */
.ack-btn {
  padding: var(--space-1) var(--space-4);
  font-size: 0.75rem;
  font-family: var(--font-body);
  font-weight: 600;
  color: var(--accent-amber);
  background: rgba(240, 160, 48, 0.08);
  border: 1px solid rgba(240, 160, 48, 0.25);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.ack-btn:hover:not(:disabled) {
  background: rgba(240, 160, 48, 0.18);
  border-color: var(--accent-amber);
}

.ack-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.acked-label {
  font-size: 0.72rem;
  color: var(--color-text-muted);
  font-weight: 500;
}

/* ====== Responsive ====== */
@media (max-width: 900px) {
  .alert-table-wrap {
    overflow-x: auto;
  }

  .alert-table {
    min-width: 680px;
  }
}
</style>
