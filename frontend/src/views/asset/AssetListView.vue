<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { get, post, del } from '@/api/http'

interface Asset {
  id: number
  name: string
  ip: string
  type: string
  status: string
  location: string
}

interface AssetType {
  id: number
  name: string
}

const assets = ref<Asset[]>([])
const assetTypes = ref<AssetType[]>([])
const loading = ref(false)
const error = ref('')

// Create form state
const showForm = ref(false)
const formLoading = ref(false)
const formError = ref('')
const formData = ref({
  name: '',
  ip: '',
  type: '',
  status: 'online',
  location: '',
})

const statusLabels: Record<string, string> = {
  online: '在线',
  offline: '离线',
  maintenance: '维护中',
  warning: '告警',
}

const statusColors: Record<string, string> = {
  online: 'var(--accent-emerald)',
  offline: 'var(--color-text-muted)',
  maintenance: 'var(--accent-amber)',
  warning: 'var(--accent-red)',
}

async function fetchAssets(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const res = await get<Asset[]>('/assets')
    assets.value = res.data || []
  } catch (err: unknown) {
    if (err instanceof Error) {
      error.value = err.message
    } else {
      error.value = '获取资产列表失败'
    }
  } finally {
    loading.value = false
  }
}

async function fetchAssetTypes(): Promise<void> {
  try {
    const res = await get<AssetType[]>('/asset-types')
    assetTypes.value = res.data || []
  } catch {
    // Silently fail - asset types are optional for the form
  }
}

async function handleCreate(): Promise<void> {
  if (!formData.value.name || !formData.value.ip || !formData.value.type) {
    formError.value = '请填写必填字段（名称、IP、类型）'
    return
  }

  formLoading.value = true
  formError.value = ''

  try {
    await post<Asset>('/assets', formData.value)
    showForm.value = false
    formData.value = { name: '', ip: '', type: '', status: 'online', location: '' }
    await fetchAssets()
  } catch (err: unknown) {
    if (err instanceof Error) {
      formError.value = err.message
    } else {
      formError.value = '创建资产失败'
    }
  } finally {
    formLoading.value = false
  }
}

async function handleDelete(id: number): Promise<void> {
  try {
    await del<null>(`/assets/${id}`)
    await fetchAssets()
  } catch {
    // Silently fail on delete
  }
}

function openForm(): void {
  showForm.value = true
  formError.value = ''
}

function closeForm(): void {
  showForm.value = false
  formError.value = ''
  formData.value = { name: '', ip: '', type: '', status: 'online', location: '' }
}

onMounted(() => {
  fetchAssets()
  fetchAssetTypes()
})
</script>

<template>
  <div class="asset-page">
    <header class="page-header">
      <div class="page-header-left">
        <h1 class="page-title">
          <span class="title-icon">⬡</span>
          资产管理
        </h1>
        <span class="asset-count">共 {{ assets.length }} 项资产</span>
      </div>
      <button class="btn-create" @click="openForm">
        <span class="btn-icon">＋</span>
        新增资产
      </button>
    </header>

    <!-- Error state -->
    <div v-if="error" class="error-banner">
      <span class="error-icon">▲</span>
      {{ error }}
      <button class="retry-btn" @click="fetchAssets">重试</button>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="loading-state">
      <span class="loading-spinner"></span>
      <span>加载中...</span>
    </div>

    <!-- Asset table -->
    <div v-if="!loading && !error" class="table-container">
      <table class="asset-table">
        <thead>
          <tr>
            <th>名称</th>
            <th>IP 地址</th>
            <th>类型</th>
            <th>状态</th>
            <th>位置</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="assets.length === 0">
            <td colspan="6" class="empty-cell">暂无资产数据</td>
          </tr>
          <tr v-for="asset in assets" :key="asset.id">
            <td>
              <span class="asset-name">{{ asset.name }}</span>
            </td>
            <td>
              <code class="asset-ip">{{ asset.ip }}</code>
            </td>
            <td>
              <span class="asset-type-badge">{{ asset.type }}</span>
            </td>
            <td>
              <span class="status-dot" :style="{ background: statusColors[asset.status] || 'var(--color-text-muted)' }"></span>
              <span class="status-text">{{ statusLabels[asset.status] || asset.status }}</span>
            </td>
            <td>
              <span class="asset-location">{{ asset.location || '—' }}</span>
            </td>
            <td>
              <button class="btn-delete" @click="handleDelete(asset.id)" title="删除">✕</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Create form modal -->
    <Teleport to="body">
      <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
        <div class="modal-content">
          <header class="modal-header">
            <h2 class="modal-title">新增资产</h2>
            <button class="modal-close" @click="closeForm">✕</button>
          </header>

          <form class="modal-form" @submit.prevent="handleCreate">
            <div class="form-row">
              <label class="form-label">
                资产名称 <span class="required">*</span>
              </label>
              <input
                v-model="formData.name"
                type="text"
                class="form-input"
                placeholder="例如：PROD-DB-Master-01"
                :disabled="formLoading"
              />
            </div>

            <div class="form-row">
              <label class="form-label">
                IP 地址 <span class="required">*</span>
              </label>
              <input
                v-model="formData.ip"
                type="text"
                class="form-input"
                placeholder="例如：10.0.1.100"
                :disabled="formLoading"
              />
            </div>

            <div class="form-row">
              <label class="form-label">
                资产类型 <span class="required">*</span>
              </label>
              <select v-model="formData.type" class="form-input form-select" :disabled="formLoading">
                <option value="" disabled>请选择类型</option>
                <option v-for="at in assetTypes" :key="at.id" :value="at.name">{{ at.name }}</option>
                <option value="服务器">服务器</option>
                <option value="交换机">交换机</option>
                <option value="路由器">路由器</option>
                <option value="存储">存储</option>
                <option value="防火墙">防火墙</option>
              </select>
            </div>

            <div class="form-row">
              <label class="form-label">状态</label>
              <select v-model="formData.status" class="form-input form-select" :disabled="formLoading">
                <option value="online">在线</option>
                <option value="offline">离线</option>
                <option value="maintenance">维护中</option>
                <option value="warning">告警</option>
              </select>
            </div>

            <div class="form-row">
              <label class="form-label">位置</label>
              <input
                v-model="formData.location"
                type="text"
                class="form-input"
                placeholder="例如：机房A-3层-机柜05"
                :disabled="formLoading"
              />
            </div>

            <div v-if="formError" class="form-error">
              <span class="error-icon">▲</span>
              {{ formError }}
            </div>

            <div class="form-actions">
              <button type="button" class="btn-cancel" @click="closeForm" :disabled="formLoading">取消</button>
              <button type="submit" class="btn-submit" :disabled="formLoading">
                <span v-if="formLoading" class="btn-spinner"></span>
                <span v-else>创建</span>
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.asset-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-wrap: wrap;
  gap: var(--space-4);
}

.page-header-left {
  display: flex;
  align-items: baseline;
  gap: var(--space-4);
}

.page-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 1.2rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.title-icon {
  color: var(--accent-cyan);
  font-size: 1rem;
}

.asset-count {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.btn-create {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-4);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.12) 0%, rgba(0, 212, 255, 0.04) 100%);
  border: 1px solid rgba(0, 212, 255, 0.3);
  border-radius: var(--radius-md);
  color: var(--accent-cyan);
  font-family: var(--font-body);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-create:hover {
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.2) 0%, rgba(0, 212, 255, 0.08) 100%);
  box-shadow: 0 0 12px rgba(0, 212, 255, 0.15);
}

.btn-icon {
  font-size: 1rem;
  line-height: 1;
}

/* ====== Error / Loading ====== */
.error-banner {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  background: rgba(240, 72, 72, 0.06);
  border: 1px solid rgba(240, 72, 72, 0.2);
  border-radius: var(--radius-md);
  color: var(--accent-red);
  font-size: 0.85rem;
}

.error-icon {
  font-size: 0.7rem;
  flex-shrink: 0;
}

.retry-btn {
  margin-left: auto;
  padding: var(--space-1) var(--space-3);
  background: rgba(240, 72, 72, 0.1);
  border: 1px solid rgba(240, 72, 72, 0.3);
  border-radius: var(--radius-sm);
  color: var(--accent-red);
  font-size: 0.75rem;
  cursor: pointer;
  font-family: var(--font-body);
  transition: background var(--transition-fast);
}

.retry-btn:hover {
  background: rgba(240, 72, 72, 0.2);
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-12) 0;
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(0, 212, 255, 0.2);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ====== Table ====== */
.table-container {
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  box-shadow: var(--shadow-card);
}

.asset-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.asset-table th {
  padding: var(--space-3) var(--space-4);
  text-align: left;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
}

.asset-table td {
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
}

.asset-table tbody tr {
  transition: background var(--transition-fast);
}

.asset-table tbody tr:hover {
  background: rgba(0, 212, 255, 0.03);
}

.asset-table tbody tr:last-child td {
  border-bottom: none;
}

.empty-cell {
  text-align: center;
  padding: var(--space-10) var(--space-4);
  color: var(--color-text-muted);
  font-size: 0.85rem;
}

.asset-name {
  color: var(--color-text-primary);
  font-weight: 500;
}

.asset-ip {
  font-family: var(--font-display);
  font-size: 0.8rem;
  color: var(--accent-cyan);
  background: rgba(0, 212, 255, 0.06);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
}

.asset-type-badge {
  font-size: 0.75rem;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: rgba(14, 165, 160, 0.1);
  color: var(--accent-teal);
  border: 1px solid rgba(14, 165, 160, 0.15);
}

.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  margin-right: var(--space-2);
  vertical-align: middle;
}

.status-text {
  font-size: 0.8rem;
  vertical-align: middle;
}

.asset-location {
  font-size: 0.8rem;
}

.btn-delete {
  padding: var(--space-1) var(--space-2);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--color-text-muted);
  cursor: pointer;
  font-size: 0.75rem;
  transition: all var(--transition-fast);
}

.btn-delete:hover {
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.08);
  border-color: rgba(240, 72, 72, 0.2);
}

/* ====== Modal ====== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(6, 10, 18, 0.8);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: var(--space-4);
}

.modal-content {
  width: 100%;
  max-width: 480px;
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 40px rgba(0, 0, 0, 0.5), var(--shadow-glow-cyan);
  overflow: hidden;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--color-border);
}

.modal-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.modal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  font-size: 0.85rem;
  transition: all var(--transition-fast);
}

.modal-close:hover {
  background: rgba(240, 72, 72, 0.08);
  color: var(--accent-red);
}

.modal-form {
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.required {
  color: var(--accent-red);
}

.form-input {
  padding: var(--space-3) var(--space-4);
  background: var(--color-bg-deep);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: 0.85rem;
  outline: none;
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.form-input::placeholder {
  color: var(--color-text-muted);
}

.form-input:focus {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.form-select {
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='%238899b4' viewBox='0 0 16 16'%3E%3Cpath d='M8 11L3 6h10z'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 12px center;
  padding-right: var(--space-8);
}

.form-select option {
  background: var(--color-bg-deep);
  color: var(--color-text-primary);
}

.form-error {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  background: rgba(240, 72, 72, 0.08);
  border: 1px solid rgba(240, 72, 72, 0.2);
  border-radius: var(--radius-md);
  color: var(--accent-red);
  font-size: 0.8rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding-top: var(--space-2);
}

.btn-cancel {
  padding: var(--space-2) var(--space-5);
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-family: var(--font-body);
  font-size: 0.8rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-cancel:hover {
  border-color: var(--color-text-muted);
  color: var(--color-text-primary);
}

.btn-cancel:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-submit {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-2) var(--space-5);
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.15) 0%, rgba(0, 212, 255, 0.05) 100%);
  border: 1px solid var(--accent-cyan);
  border-radius: var(--radius-md);
  color: var(--accent-cyan);
  font-family: var(--font-display);
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: 0.06em;
  cursor: pointer;
  transition: all var(--transition-fast);
  min-width: 80px;
}

.btn-submit:hover:not(:disabled) {
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.25) 0%, rgba(0, 212, 255, 0.1) 100%);
  box-shadow: 0 0 16px rgba(0, 212, 255, 0.2);
}

.btn-submit:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-spinner {
  width: 14px;
  height: 14px;
  border: 2px solid rgba(0, 212, 255, 0.3);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

/* ====== Responsive ====== */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .table-container {
    overflow-x: auto;
  }

  .asset-table {
    min-width: 600px;
  }
}
</style>
