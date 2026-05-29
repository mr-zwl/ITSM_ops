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
  asset_name: string
  asset_ip: string
  metric_code: string
  value: number
  collected_at: string
}

interface AssetMetrics {
  assetId: number
  assetName: string
  assetIp: string
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

    const first = dataPoints[0]!
    result.push({
      assetId,
      assetName: first.asset_name || `资产 ${assetId}`,
      assetIp: first.asset_ip || '-',
      metrics,
    })
  }

  return result
})

// Search
const searchQuery = ref('')

const filteredGroupedAssets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return groupedByAsset.value
  return groupedByAsset.value.filter(g => {
    return g.assetName.toLowerCase().includes(q)
      || g.assetIp.toLowerCase().includes(q)
  })
})

function metricColor(percentage: number): string {
  if (percentage >= 95) return 'var(--accent-red)'
  if (percentage >= 80) return 'var(--accent-amber)'
  return 'var(--accent-emerald)'
}

function metricBarClass(percentage: number): string {
  if (percentage >= 95) return 'bar-fill--red'
  if (percentage >= 80) return 'bar-fill--amber'
  return 'bar-fill--green'
}

async function fetchData(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const [defsRes, dataRes] = await Promise.all([
      get<MetricDef[]>('/metrics'),
      get<MetricDataPoint[]>('/metric-data?limit=500'),
    ])
    metricDefs.value = defsRes.data || []
    metricData.value = dataRes.data || []
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : '加载指标数据失败'
  } finally {
    loading.value = false
  }
}

function formatTime(iso: string): string {
  return new Date(iso).toLocaleString('zh-CN', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

onMounted(fetchData)
</script>

<template>
  <div class="metric-page">
    <div class="page-header">
      <div>
        <h1 class="page-title">指标监控</h1>
        <p class="page-desc">实时监控资产运行指标</p>
      </div>
      <div class="search-box">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            class="search-input"
            type="text"
            placeholder="搜索资产名称、IP..."
          />
          <span v-if="searchQuery" class="search-count">{{ filteredGroupedAssets.length }} / {{ groupedByAsset.length }}</span>
          <button v-if="searchQuery" class="search-clear" @click="searchQuery = ''">✕</button>
        </div>
        <button class="btn-secondary" @click="fetchData">🔄 刷新</button>
    </div>

    <div v-if="error" class="form-error">{{ error }}</div>
    <div v-if="loading" class="loading-hint">加载中...</div>

    <div v-else class="asset-metric-grid">
      <div v-for="asset in filteredGroupedAssets" :key="asset.assetId" class="metric-card">
        <div class="metric-card-header">
          <span class="metric-asset-name">💻 {{ asset.assetName }}</span>
          <span class="metric-asset-ip">{{ asset.assetIp }}</span>
        </div>

        <div class="metric-list">
          <div v-for="m in asset.metrics" :key="m.code" class="metric-item">
            <div class="metric-item-header">
              <span class="metric-name">{{ m.name }}</span>
              <span class="metric-value" :style="{ color: metricColor(m.percentage) }">
                {{ m.value.toFixed(1) }}{{ m.unit }}
              </span>
            </div>
            <div class="metric-bar-track">
              <div class="metric-bar-fill" :class="metricBarClass(m.percentage)" :style="{ width: m.percentage + '%' }"></div>
            </div>
            <span class="metric-time">{{ formatTime(m.collectedAt) }}</span>
          </div>
        </div>
      </div>

      <div v-if="filteredGroupedAssets.length === 0" class="empty-state">
        <span class="empty-icon">📊</span>
        <p>暂无指标数据</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.metric-page {
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

.btn-secondary {
  padding: var(--space-2) var(--space-4);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-base);
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
}

.btn-secondary:hover {
  border-color: var(--color-text-muted);
  color: var(--color-text-primary);
}

.form-error {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--accent-pink);
  color: var(--accent-red);
  font-size: 0.85rem;
  font-weight: 500;
}

.asset-metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
  gap: var(--space-4);
}

.metric-card {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.metric-card:hover {
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-2px);
}

.metric-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.metric-asset-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.metric-asset-ip {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.metric-list {
  padding: var(--space-4) var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.metric-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.metric-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.metric-name {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.metric-value {
  font-size: 0.9rem;
  font-weight: 700;
}

.metric-bar-track {
  height: 6px;
  background: var(--color-border);
  border-radius: 3px;
  overflow: hidden;
}

.metric-bar-fill {
  height: 100%;
  border-radius: 3px;
  transition: width var(--transition-base);
}

.bar-fill--green { background: var(--accent-emerald); }
.bar-fill--amber { background: var(--accent-amber); }
.bar-fill--red { background: var(--accent-red); }

.metric-time {
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.empty-state {
  grid-column: 1 / -1;
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

/* ====== Search ====== */
.search-box {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  background: var(--color-bg-base);
  border: 1.5px solid var(--color-border-subtle);
  border-radius: 100px;
  min-width: 260px;
  transition: all var(--transition-fast);
}

.search-box:focus-within {
  border-color: var(--accent-red);
  box-shadow: 0 0 0 3px rgba(255, 36, 66, 0.08);
}

.search-icon {
  font-size: 0.9rem;
  flex-shrink: 0;
  opacity: 0.6;
}

.search-input {
  flex: 1;
  border: none;
  outline: none;
  font-size: 0.9rem;
  color: var(--color-text-primary);
  background: transparent;
  font-family: var(--font-body);
  min-width: 0;
}

.search-input::placeholder {
  color: var(--color-text-muted);
}

.search-count {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  white-space: nowrap;
  flex-shrink: 0;
}

.search-clear {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: none;
  border-radius: 50%;
  background: var(--color-border);
  color: var(--color-text-muted);
  font-size: 0.7rem;
  cursor: pointer;
  flex-shrink: 0;
  transition: all var(--transition-fast);
}

.search-clear:hover {
  background: var(--accent-pink);
  color: var(--accent-red);
}

</style>
