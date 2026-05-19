<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { get } from '@/api/http'

interface MetricDef {
  id: number
  code: string
  name: string
  unit: string
}

interface MetricDataPoint {
  asset_id: number
  metric_code: string
  value: number
  collected_at: string
}

interface AssetMetrics {
  assetId: number
  metrics: {
    code: string
    name: string
    unit: string
    value: number
    collectedAt: string
    percentage: number
  }[]
}

const loading = ref(true)
const error = ref('')
const metricDefs = ref<MetricDef[]>([])
const metricData = ref<MetricDataPoint[]>([])

const metricDefMap = computed(() => {
  const map = new Map<string, MetricDef>()
  for (const def of metricDefs.value) {
    map.set(def.code, def)
  }
  return map
})

const groupedByAsset = computed<AssetMetrics[]>(() => {
  const groups = new Map<number, MetricDataPoint[]>()

  for (const dp of metricData.value) {
    const existing = groups.get(dp.asset_id)
    if (existing) {
      existing.push(dp)
    } else {
      groups.set(dp.asset_id, [dp])
    }
  }

  const result: AssetMetrics[] = []
  for (const [assetId, dataPoints] of groups) {
    const latestByMetric = new Map<string, MetricDataPoint>()
    for (const dp of dataPoints) {
      const current = latestByMetric.get(dp.metric_code)
      if (!current || new Date(dp.collected_at) > new Date(current.collected_at)) {
        latestByMetric.set(dp.metric_code, dp)
      }
    }

    const metrics: AssetMetrics['metrics'] = []
    for (const [code, dp] of latestByMetric) {
      const def = metricDefMap.value.get(code)
      metrics.push({
        code,
        name: def?.name ?? code,
        unit: def?.unit ?? '',
        value: dp.value,
        collectedAt: dp.collected_at,
        percentage: Math.min(Math.max(dp.value, 0), 100),
      })
    }

    result.push({ assetId, metrics })
  }

  return result
})

function gaugeColor(pct: number): string {
  if (pct > 80) return 'var(--accent-red)'
  if (pct > 60) return 'var(--accent-amber)'
  return 'var(--accent-emerald)'
}

function gaugeGlow(pct: number): string {
  if (pct > 80) return '0 0 10px rgba(240, 72, 72, 0.4)'
  if (pct > 60) return '0 0 10px rgba(240, 160, 48, 0.4)'
  return '0 0 10px rgba(45, 212, 160, 0.4)'
}

function gaugeGradient(pct: number): string {
  if (pct > 80) return 'linear-gradient(90deg, var(--accent-amber), var(--accent-red))'
  if (pct > 60) return 'linear-gradient(90deg, var(--accent-emerald), var(--accent-amber))'
  return 'linear-gradient(90deg, var(--accent-cyan), var(--accent-emerald))'
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

async function fetchData(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const [defsRes, dataRes] = await Promise.all([
      get<MetricDef[]>('/metrics'),
      get<MetricDataPoint[]>('/metric-data?limit=50'),
    ])
    metricDefs.value = defsRes.data
    metricData.value = dataRes.data
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载指标数据失败'
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="metric-page">
    <header class="page-header">
      <h1 class="page-title">
        <span class="title-icon" style="color: var(--accent-cyan)">◈</span>
        指标监控
      </h1>
      <span class="page-subtitle">实时资产指标概览</span>
    </header>

    <div v-if="loading" class="state-panel">
      <div class="loading-ring"></div>
      <span class="state-text">加载中...</span>
    </div>

    <div v-else-if="error" class="state-panel state-panel--error">
      <span class="state-icon">⚠</span>
      <span class="state-text">{{ error }}</span>
      <button class="retry-btn" @click="fetchData">重试</button>
    </div>

    <div v-else-if="groupedByAsset.length === 0" class="state-panel">
      <span class="state-icon">◇</span>
      <span class="state-text">暂无指标数据</span>
    </div>

    <section v-else class="asset-grid">
      <article v-for="group in groupedByAsset" :key="group.assetId" class="asset-card">
        <div class="asset-card__header">
          <span class="asset-id-badge">资产 #{{ group.assetId }}</span>
          <span class="metric-count">{{ group.metrics.length }} 项指标</span>
        </div>

        <div class="gauge-list">
          <div v-for="m in group.metrics" :key="m.code" class="gauge-row">
            <div class="gauge-row__info">
              <span class="gauge-name">{{ m.name }}</span>
              <span class="gauge-value" :style="{ color: gaugeColor(m.percentage) }">
                {{ m.value.toFixed(1) }}<span class="gauge-unit">{{ m.unit }}</span>
              </span>
            </div>
            <div class="gauge-track">
              <div
                class="gauge-fill"
                :style="{
                  width: m.percentage + '%',
                  background: gaugeGradient(m.percentage),
                  boxShadow: gaugeGlow(m.percentage),
                }"
              ></div>
            </div>
            <span class="gauge-pct" :style="{ color: gaugeColor(m.percentage) }">
              {{ Math.round(m.percentage) }}%
            </span>
          </div>
        </div>

        <div class="asset-card__footer">
          <span class="update-time">更新于 {{ formatTime(group.metrics[0]?.collectedAt ?? '') }}</span>
        </div>
      </article>
    </section>
  </div>
</template>

<style scoped>
.metric-page {
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

/* ====== Asset Grid ====== */
.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: var(--space-5);
}

.asset-card {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-5) var(--space-6);
  box-shadow: var(--shadow-card);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
  transition: border-color var(--transition-fast);
  position: relative;
  overflow: hidden;
}

.asset-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, var(--accent-cyan), transparent);
  opacity: 0.4;
}

.asset-card:hover {
  border-color: var(--color-border-subtle);
}

.asset-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.asset-id-badge {
  font-family: var(--font-display);
  font-size: 0.85rem;
  font-weight: 700;
  color: var(--accent-cyan);
  letter-spacing: 0.04em;
}

.metric-count {
  font-size: 0.7rem;
  color: var(--color-text-muted);
  background: var(--color-bg-deep);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

/* ====== Gauge Rows ====== */
.gauge-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.gauge-row {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.gauge-row__info {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  min-width: 120px;
  flex-shrink: 0;
}

.gauge-name {
  font-size: 0.78rem;
  color: var(--color-text-secondary);
}

.gauge-value {
  font-family: var(--font-display);
  font-size: 0.8rem;
  font-weight: 700;
  margin-left: auto;
}

.gauge-unit {
  font-size: 0.65rem;
  font-weight: 400;
  opacity: 0.7;
  margin-left: 2px;
}

.gauge-track {
  flex: 1;
  height: 10px;
  background: var(--color-bg-deep);
  border-radius: var(--radius-sm);
  overflow: hidden;
  position: relative;
}

.gauge-track::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: repeating-linear-gradient(
    90deg,
    transparent,
    transparent 19%,
    rgba(255, 255, 255, 0.02) 19%,
    rgba(255, 255, 255, 0.02) 20%
  );
  pointer-events: none;
}

.gauge-fill {
  height: 100%;
  border-radius: var(--radius-sm);
  transition: width var(--transition-base);
  position: relative;
}

.gauge-pct {
  font-family: var(--font-display);
  font-size: 0.75rem;
  font-weight: 600;
  width: 36px;
  text-align: right;
  flex-shrink: 0;
}

.asset-card__footer {
  padding-top: var(--space-2);
  border-top: 1px solid var(--color-border-subtle);
}

.update-time {
  font-size: 0.68rem;
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
}

/* ====== Responsive ====== */
@media (max-width: 860px) {
  .asset-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 480px) {
  .gauge-row__info {
    flex-direction: column;
    min-width: 80px;
    gap: 0;
  }

  .gauge-value {
    margin-left: 0;
  }
}
</style>
