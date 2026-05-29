<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick, computed } from 'vue'
import { get, post, del } from '@/api/http'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

interface Asset {
  id: number
  asset_type_id: number
  name: string
  ip: string
  type: string
  status: string
  location: string
  os_type: string
  ssh_user: string
  ssh_password: string
  ssh_port: string
  rdp_user: string
  rdp_port: string
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
const showInstall = ref(false)
const installGuide = ref<{asset_id: number; asset_name: string; os_type: string; endpoint: string; commands: Record<string, string>} | null>(null)
const copiedKey = ref('')
const formData = ref({
  name: '',
  ip: '',
  asset_type_id: null as number | null,
  status: 'online',
  os_type: 'linux',
  location: '',
  ssh_user: '',
  ssh_password: '',
  ssh_port: '22',
  rdp_user: '',
  rdp_port: '3389',
})

// SSH terminal state
const showTerminal = ref(false)
const terminalAsset = ref<Asset | null>(null)
const terminalEl = ref<HTMLElement | null>(null)
const termConnected = ref(false)
const termConnecting = ref(false)
const showCredForm = ref(false)
const credForm = ref({ host: '', port: '22', user: '', password: '' })

let term: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

const statusLabels: Record<string, string> = {
  online: '在线',
  offline: '离线',
  maintenance: '维护中',
  warning: '告警',
}

// Search
const searchQuery = ref('')

const filteredAssets = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return assets.value
  return assets.value.filter(a => {
    return a.name.toLowerCase().includes(q)
      || a.ip.toLowerCase().includes(q)
      || (a.location && a.location.toLowerCase().includes(q))
      || (a.os_type && a.os_type.toLowerCase().includes(q))
  })
})

// statusColors removed - using CSS classes instead

function assetTypeName(id: number): string {
  const found = assetTypes.value.find(at => at.id === id)
  return found ? found.name : String(id)
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
  if (!formData.value.name || !formData.value.ip || !formData.value.asset_type_id) {
    formError.value = '请填写必填字段（名称、IP、类型）'
    return
  }

  formLoading.value = true
  formError.value = ''

  try {
    const res = await post<Asset>('/assets', formData.value)
    showForm.value = false
    // Fetch install guide
    try {
      const guideRes = await get<any>(`/assets/${res.data.id}/install`)
      installGuide.value = guideRes.data
      showInstall.value = true
    } catch { /* ignore */ }
    formData.value = { name: '', ip: '', asset_type_id: null as number | null, status: 'online', os_type: 'linux', location: '', ssh_user: '', ssh_password: '', ssh_port: '22', rdp_user: '', rdp_port: '3389' }
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

function copyCommand(key: string): void {
  if (!installGuide.value) return
  const cmd = installGuide.value.commands[key]
  if (cmd) {
    navigator.clipboard.writeText(cmd)
    copiedKey.value = key
    setTimeout(() => { copiedKey.value = '' }, 2000)
  }
}

function closeForm(): void {
  showForm.value = false
  formError.value = ''
  formData.value = { name: '', ip: '', asset_type_id: null as number | null, status: 'online', os_type: 'linux', location: '', ssh_user: '', ssh_password: '', ssh_port: '22', rdp_user: '', rdp_port: '3389' }
}

// ====== Remote access ======

function handleRemote(asset: Asset): void {
  if (asset.os_type === 'windows') {
    window.location.href = `/api/v1/assets/${asset.id}/rdp`
  } else {
    openTerminal(asset)
  }
}

function openTerminal(asset: Asset): void {
  terminalAsset.value = asset
  showTerminal.value = true
  termConnected.value = false
  termConnecting.value = false
  showCredForm.value = false

  // If asset has credentials, auto-connect
  if (asset.ssh_user && asset.ssh_password) {
    credForm.value = {
      host: asset.ip,
      port: asset.ssh_port || '22',
      user: asset.ssh_user,
      password: asset.ssh_password,
    }
    nextTick(() => connectSSH())
  } else {
    // Show credential form
    showCredForm.value = true
    credForm.value = {
      host: asset.ip,
      port: asset.ssh_port || '22',
      user: '',
      password: '',
    }
  }
}

function initTerminal(): void {
  if (!terminalEl.value) return

  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: '"SF Mono", "Cascadia Code", "Fira Code", monospace',
    theme: {
      background: '#060a12',
      foreground: '#e8edf5',
      cursor: '#00d4ff',
      selectionBackground: 'rgba(0, 212, 255, 0.25)',
      black: '#0b1120',
      red: '#f04848',
      green: '#2dd4a0',
      yellow: '#f0a030',
      blue: '#00d4ff',
      magenta: '#c084fc',
      cyan: '#0ea5a0',
      white: '#e8edf5',
      brightBlack: '#4d6282',
      brightRed: '#f87171',
      brightGreen: '#4ade80',
      brightYellow: '#fbbf24',
      brightBlue: '#38bdf8',
      brightMagenta: '#d8b4fe',
      brightCyan: '#22d3ee',
      brightWhite: '#f1f5f9',
    },
    allowProposedApi: true,
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  term.open(terminalEl.value)
  fitAddon.fit()

  term.onData((data: string) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(new TextEncoder().encode(data))
    }
  })

  term.onResize(({ rows, cols }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ resize: { rows, cols } }))
    }
  })

  window.addEventListener('resize', handleResize)
}

function handleResize(): void {
  if (fitAddon && term) {
    fitAddon.fit()
  }
}

function connectSSH(): void {
  if (!terminalAsset.value) return

  showCredForm.value = false
  termConnecting.value = true
  termConnected.value = false

  // Init terminal if not yet
  nextTick(() => {
    if (!term) {
      initTerminal()
    }

    const asset = terminalAsset.value!
    const wsProto = location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${wsProto}//${location.host}/api/v1/ssh?asset_id=${asset.id}`

    ws = new WebSocket(wsUrl)
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      termConnecting.value = false
      termConnected.value = true

      // Send credentials as first message
      const creds = JSON.stringify({
        host: credForm.value.host || asset.ip,
        port: parseInt(credForm.value.port || '22', 10),
        user: credForm.value.user,
        password: credForm.value.password,
      })
      ws!.send(creds)

      // Fit after connection
      nextTick(() => {
        if (fitAddon && term) {
          fitAddon.fit()
        }
      })
    }

    ws.onmessage = (event: MessageEvent) => {
      if (!term) return
      if (event.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(event.data))
      } else {
        term.write(event.data)
      }
    }

    ws.onclose = () => {
      termConnecting.value = false
      termConnected.value = false
      if (term) {
        term.write('\r\n\x1b[33m--- 连接已断开 ---\x1b[0m\r\n')
      }
    }

    ws.onerror = () => {
      termConnecting.value = false
      termConnected.value = false
      if (term) {
        term.write('\r\n\x1b[31m--- 连接失败 ---\x1b[0m\r\n')
      }
    }
  })
}

function closeTerminal(): void {
  if (ws) {
    ws.close()
    ws = null
  }
  if (term) {
    term.dispose()
    term = null
  }
  fitAddon = null
  window.removeEventListener('resize', handleResize)
  showTerminal.value = false
  terminalAsset.value = null
  termConnected.value = false
  termConnecting.value = false
  showCredForm.value = false
}

onMounted(() => {
  fetchAssets()
  fetchAssetTypes()
})

onBeforeUnmount(() => {
  closeTerminal()
})
</script>

<template>
  <div class="asset-page">
    <!-- 页面标题 -->
    <div class="page-header">
      <div>
        <h1 class="page-title">资产管理</h1>
        <p class="page-desc">管理和监控所有 IT 资产</p>
      </div>
      <div class="search-box">
          <span class="search-icon">🔍</span>
          <input
            v-model="searchQuery"
            class="search-input"
            type="text"
            placeholder="搜索资产名称、IP、机房位置..."
            clearable
          />
          <span v-if="searchQuery" class="search-count">{{ filteredAssets.length }} / {{ assets.length }}</span>
          <button v-if="searchQuery" class="search-clear" @click="searchQuery = ''">✕</button>
        </div>
        <button class="btn-primary" @click="openForm">＋ 新增资产</button>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="form-error">{{ error }}</div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-hint">加载中...</div>

    <!-- 资产卡片网格 -->
    <div v-else class="asset-grid">
      <div v-for="asset in filteredAssets" :key="asset.id" class="asset-card">
        <div class="asset-card-header">
          <div class="asset-icon">{{ asset.os_type === 'windows' ? '🖥️' : '💻' }}</div>
          <div class="asset-main-info">
            <span class="asset-name">{{ asset.name }}</span>
            <span class="asset-ip">{{ asset.ip }}</span>
          </div>
          <span class="status-badge" :class="`status-badge--${asset.status}`">{{ statusLabels[asset.status] || asset.status }}</span>
        </div>

        <div class="asset-card-body">
          <div class="asset-meta">
            <span class="meta-item">
              <span class="meta-label">类型</span>
              <span class="meta-value">{{ assetTypeName(asset.asset_type_id) }}</span>
            </span>
            <span class="meta-item">
              <span class="meta-label">系统</span>
              <span class="meta-value xhs-tag xhs-tag--blue">{{ asset.os_type || 'linux' }}</span>
            </span>
            <span class="meta-item" v-if="asset.location">
              <span class="meta-label">位置</span>
              <span class="meta-value">{{ asset.location }}</span>
            </span>
          </div>
        </div>

        <div class="asset-card-actions">
          <button class="btn-action btn-action--primary" @click="handleRemote(asset)">
            {{ asset.os_type === 'windows' ? '🖥️ 远程桌面' : '⌨️ SSH' }}
          </button>
          <button class="btn-action btn-action--secondary" @click="showInstall = true; installGuide = null; get<any>(`/assets/${asset.id}/install`).then(r => { if(r.code===0) installGuide = r.data as any }).catch(() => {})">📋 安装</button>
          <button class="btn-action btn-action--danger" @click="handleDelete(asset.id)">🗑️</button>
        </div>
      </div>

      <div v-if="filteredAssets.length === 0" class="empty-state">
        <span class="empty-icon">📦</span>
        <p>暂无资产，点击上方按钮添加</p>
      </div>
    </div>

    <!-- 新增资产弹窗 -->
    <div v-if="showForm" class="modal-overlay" @click.self="closeForm">
      <div class="modal-card">
        <div class="modal-header">
          <h2 class="modal-title">新增资产</h2>
          <button class="modal-close" @click="closeForm">✕</button>
        </div>

        <form class="modal-body" @submit.prevent="handleCreate">
          <div v-if="formError" class="form-error">{{ formError }}</div>

          <div class="form-group">
            <label class="form-label">名称 *</label>
            <input v-model="formData.name" class="form-input" placeholder="如：Web-Server-01" />
          </div>

          <div class="form-group">
            <label class="form-label">IP 地址 *</label>
            <input v-model="formData.ip" class="form-input" placeholder="如：192.168.1.100" />
          </div>

          <div class="form-row">
            <div class="form-group">
              <label class="form-label">类型 *</label>
              <select v-model="formData.asset_type_id" class="form-input">
                <option :value="null" disabled>选择类型</option>
                <option v-for="at in assetTypes" :key="at.id" :value="at.id">{{ at.name }}</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">操作系统</label>
              <select v-model="formData.os_type" class="form-input">
                <option value="linux">Linux</option>
                <option value="windows">Windows</option>
              </select>
            </div>
          </div>

          <div class="form-group">
            <label class="form-label">位置</label>
            <input v-model="formData.location" class="form-input" placeholder="如：北京机房" />
          </div>

          <div class="form-section-title">SSH 连接</div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">用户名</label>
              <input v-model="formData.ssh_user" class="form-input" placeholder="root" />
            </div>
            <div class="form-group">
              <label class="form-label">端口</label>
              <input v-model="formData.ssh_port" class="form-input" placeholder="22" />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">密码</label>
            <input v-model="formData.ssh_password" type="password" class="form-input" placeholder="SSH 密码" />
          </div>

          <div class="form-section-title" v-if="formData.os_type === 'windows'">RDP 连接</div>
          <template v-if="formData.os_type === 'windows'">
            <div class="form-row">
              <div class="form-group">
                <label class="form-label">用户名</label>
                <input v-model="formData.rdp_user" class="form-input" placeholder="Administrator" />
              </div>
              <div class="form-group">
                <label class="form-label">端口</label>
                <input v-model="formData.rdp_port" class="form-input" placeholder="3389" />
              </div>
            </div>
          </template>

          <div class="modal-footer">
            <button type="button" class="btn-secondary" @click="closeForm">取消</button>
            <button type="submit" class="btn-primary" :disabled="formLoading">
              {{ formLoading ? '创建中...' : '创建' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 安装指南弹窗 -->
    <div v-if="showInstall && installGuide" class="modal-overlay" @click.self="showInstall = false">
      <div class="modal-card modal-card--wide">
        <div class="modal-header">
          <h2 class="modal-title">📋 安装指南 - {{ installGuide.asset_name }}</h2>
          <button class="modal-close" @click="showInstall = false">✕</button>
        </div>
        <div class="modal-body">
          <div class="install-section">
            <div class="install-label">Endpoint</div>
            <div class="install-endpoint">{{ installGuide.endpoint }}</div>
          </div>
          <div v-for="(cmd, key) in installGuide.commands" :key="key" class="install-section">
            <div class="install-label">{{ key === 'linux' ? '🚀 Linux 一键安装' : key === 'windows' ? '🚀 Windows 一键安装' : key === 'linux_manual' ? '🔧 Linux 手动安装' : '🔧 Windows 手动安装' }}</div>
            <div class="install-cmd">
              <pre>{{ cmd }}</pre>
              <button class="btn-copy" @click="copyCommand(key)">{{ copiedKey === key ? '✓ 已复制' : '复制' }}</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- SSH 终端弹窗 -->
    <div v-if="showTerminal" class="modal-overlay modal-overlay--dark" @click.self="closeTerminal">
      <div class="modal-card modal-card--terminal">
        <div class="modal-header modal-header--dark">
          <h2 class="modal-title">
            ⌨️ SSH - {{ terminalAsset?.name }}
            <span v-if="termConnected" class="conn-status conn-status--on">已连接</span>
            <span v-else-if="termConnecting" class="conn-status conn-status--connecting">连接中...</span>
            <span v-else class="conn-status conn-status--off">未连接</span>
          </h2>
          <button class="modal-close" @click="closeTerminal">✕</button>
        </div>

        <!-- SSH 凭证表单 -->
        <div v-if="showCredForm" class="cred-form">
          <div class="form-group">
            <label class="form-label">主机</label>
            <input v-model="credForm.host" class="form-input" placeholder="目标主机 IP" />
          </div>
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">端口</label>
              <input v-model="credForm.port" class="form-input" placeholder="22" />
            </div>
            <div class="form-group">
              <label class="form-label">用户名</label>
              <input v-model="credForm.user" class="form-input" placeholder="root" />
            </div>
          </div>
          <div class="form-group">
            <label class="form-label">密码</label>
            <input v-model="credForm.password" type="password" class="form-input" placeholder="SSH 密码" />
          </div>
          <button class="btn-primary" @click="connectSSH">连接</button>
        </div>

        <!-- 终端 -->
        <div v-if="!showCredForm" class="terminal-container" ref="terminalEl"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.asset-page {
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

.btn-primary {
  padding: var(--space-2) var(--space-5);
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-red);
  color: white;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
  white-space: nowrap;
}

.btn-primary:hover:not(:disabled) {
  background: #E6203C;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 36, 66, 0.25);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  padding: var(--space-2) var(--space-5);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-base);
  color: var(--color-text-secondary);
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
}

.btn-secondary:hover {
  border-color: var(--color-text-muted);
  color: var(--color-text-primary);
}

/* ====== Asset Grid ====== */
.asset-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-4);
}

.asset-card {
  background: var(--color-bg-base);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-card);
  overflow: hidden;
  transition: all var(--transition-fast);
}

.asset-card:hover {
  box-shadow: var(--shadow-card-hover);
  transform: translateY(-2px);
}

.asset-card-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.asset-icon {
  font-size: 1.8rem;
  line-height: 1;
  flex-shrink: 0;
}

.asset-main-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.asset-name {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.asset-ip {
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

.status-badge {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 100px;
  white-space: nowrap;
}

.status-badge--online {
  color: var(--accent-emerald);
  background: var(--accent-emerald-light);
}

.status-badge--offline {
  color: var(--color-text-muted);
  background: var(--color-bg-elevated);
}

.status-badge--maintenance {
  color: var(--accent-amber);
  background: var(--accent-amber-light);
}

.status-badge--warning {
  color: var(--accent-red);
  background: var(--accent-pink);
}

/* ====== Card Body ====== */
.asset-card-body {
  padding: var(--space-4) var(--space-5);
}

.asset-meta {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.meta-label {
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.meta-value {
  font-size: 0.85rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

/* ====== Card Actions ====== */
.asset-card-actions {
  display: flex;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-5);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-elevated);
}

.btn-action {
  flex: 1;
  padding: var(--space-2) var(--space-3);
  border: none;
  border-radius: var(--radius-md);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
  text-align: center;
}

.btn-action--primary {
  background: var(--accent-pink);
  color: var(--accent-red);
}

.btn-action--primary:hover {
  background: #FFE0E5;
}

.btn-action--secondary {
  background: var(--color-bg-base);
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
}

.btn-action--secondary:hover {
  border-color: var(--color-text-muted);
  color: var(--color-text-primary);
}

.btn-action--danger {
  background: var(--color-bg-base);
  color: var(--color-text-muted);
  border: 1px solid var(--color-border);
  max-width: 44px;
}

.btn-action--danger:hover {
  background: var(--accent-pink);
  color: var(--accent-red);
  border-color: var(--accent-red);
}

/* ====== Empty State ====== */
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

/* ====== Modal ====== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
  padding: var(--space-4);
}

.modal-overlay--dark {
  background: rgba(0, 0, 0, 0.6);
}

.modal-card {
  background: var(--color-bg-base);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-elevated);
  width: 100%;
  max-width: 560px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-card--wide {
  max-width: 720px;
}

.modal-card--terminal {
  max-width: 900px;
  background: #1a1a2e;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-5) var(--space-6);
  border-bottom: 1px solid var(--color-border);
}

.modal-header--dark {
  border-bottom-color: #2a2a4a;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.modal-header--dark .modal-title {
  color: #e8edf5;
}

.modal-close {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: var(--color-bg-elevated);
  color: var(--color-text-muted);
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-fast);
}

.modal-close:hover {
  background: var(--accent-pink);
  color: var(--accent-red);
}

.modal-body {
  padding: var(--space-5) var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.modal-footer {
  display: flex;
  gap: var(--space-3);
  justify-content: flex-end;
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border);
  margin-top: var(--space-2);
}

/* ====== Form ====== */
.form-error {
  padding: var(--space-3) var(--space-4);
  border-radius: var(--radius-md);
  background: var(--accent-pink);
  color: var(--accent-red);
  font-size: 0.85rem;
  font-weight: 500;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  flex: 1;
}

.form-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.form-input {
  padding: var(--space-2) var(--space-3);
  border: 1.5px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  font-size: 0.9rem;
  color: var(--color-text-primary);
  background: var(--color-bg-base);
  transition: all var(--transition-fast);
  outline: none;
  font-family: var(--font-body);
}

.form-input:focus {
  border-color: var(--accent-red);
  box-shadow: 0 0 0 3px rgba(255, 36, 66, 0.08);
}

.form-row {
  display: flex;
  gap: var(--space-3);
}

.form-section-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-primary);
  padding-top: var(--space-2);
  border-top: 1px solid var(--color-border);
}

/* ====== Install Guide ====== */
.install-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.install-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.install-endpoint {
  padding: var(--space-2) var(--space-3);
  background: var(--color-bg-elevated);
  border-radius: var(--radius-md);
  font-size: 0.85rem;
  color: var(--accent-cyan);
  word-break: break-all;
}

.install-cmd {
  position: relative;
  background: #1a1a2e;
  border-radius: var(--radius-md);
  overflow: hidden;
}

.install-cmd pre {
  padding: var(--space-4);
  font-size: 0.8rem;
  color: #e8edf5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.6;
  font-family: "SF Mono", "Cascadia Code", "Fira Code", monospace;
}

.btn-copy {
  position: absolute;
  top: var(--space-2);
  right: var(--space-2);
  padding: var(--space-1) var(--space-3);
  border: none;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.1);
  color: #aaa;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.btn-copy:hover {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

/* ====== Terminal ====== */
.cred-form {
  padding: var(--space-5);
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.terminal-container {
  height: 500px;
  padding: var(--space-2);
}

.conn-status {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 100px;
}

.conn-status--on {
  color: var(--accent-emerald);
  background: rgba(0, 200, 83, 0.15);
}

.conn-status--connecting {
  color: var(--accent-amber);
  background: rgba(255, 143, 31, 0.15);
}

.conn-status--off {
  color: #888;
  background: rgba(255, 255, 255, 0.05);
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
  min-width: 280px;
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
