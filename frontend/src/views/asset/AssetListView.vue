<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
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
const installGuide = ref<{asset_id: number; asset_name: string; os_type: string; commands: Record<string, string>} | null>(null)
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

const statusColors: Record<string, string> = {
  online: 'var(--accent-emerald)',
  offline: 'var(--color-text-muted)',
  maintenance: 'var(--accent-amber)',
  warning: 'var(--accent-red)',
}

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
            <th>ID</th>
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
            <td colspan="7" class="empty-cell">暂无资产数据</td>
          </tr>
          <tr v-for="asset in assets" :key="asset.id">
            <td>
              <code class="asset-id">{{ asset.id }}</code>
            </td>
            <td>
              <span class="asset-name">{{ asset.name }}</span>
            </td>
            <td>
              <code class="asset-ip">{{ asset.ip }}</code>
            </td>
            <td>
              <span class="asset-type-badge">{{ assetTypeName(asset.asset_type_id) }}</span>
            </td>
            <td>
              <span class="status-dot" :style="{ background: statusColors[asset.status] || 'var(--color-text-muted)' }"></span>
              <span class="status-text">{{ statusLabels[asset.status] || asset.status }}</span>
            </td>
            <td>
              <span class="asset-location">{{ asset.location || '—' }}</span>
            </td>
            <td>
              <div class="action-btns">
                <button
                  class="btn-remote"
                  @click="handleRemote(asset)"
                  :title="asset.os_type === 'windows' ? 'RDP 远程桌面' : 'SSH 远程终端'"
                >
                  <span v-if="asset.os_type === 'windows'">⬤</span>
                  <span v-else>⌨</span>
                </button>
                <button class="btn-delete" @click="handleDelete(asset.id)" title="删除">✕</button>
              </div>
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
              <select v-model="formData.asset_type_id" class="form-input form-select" :disabled="formLoading">
                <option :value="0" disabled>请选择类型</option>
                <option v-for="at in assetTypes" :key="at.id" :value="at.id">{{ at.name }}</option>
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
              <label class="form-label">操作系统</label>
              <select v-model="formData.os_type" class="form-input form-select" :disabled="formLoading">
                <option value="linux">Linux</option>
                <option value="windows">Windows</option>
                <option value="other">其他</option>
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

            <!-- SSH credentials -->
            <div class="form-section-title">SSH 连接凭证</div>
            <div class="form-row-inline">
              <div class="form-row form-row--flex">
                <label class="form-label">SSH用户</label>
                <input
                  v-model="formData.ssh_user"
                  type="text"
                  class="form-input"
                  placeholder="root"
                  :disabled="formLoading"
                />
              </div>
              <div class="form-row form-row--flex">
                <label class="form-label">SSH端口</label>
                <input
                  v-model="formData.ssh_port"
                  type="text"
                  class="form-input"
                  placeholder="22"
                  :disabled="formLoading"
                />
              </div>
            </div>
            <div class="form-row">
              <label class="form-label">SSH密码</label>
              <input
                v-model="formData.ssh_password"
                type="password"
                class="form-input"
                placeholder="可留空"
                :disabled="formLoading"
              />
            </div>

            <!-- RDP credentials -->
            <div class="form-section-title">RDP 连接凭证</div>
            <div class="form-row-inline">
              <div class="form-row form-row--flex">
                <label class="form-label">RDP用户</label>
                <input
                  v-model="formData.rdp_user"
                  type="text"
                  class="form-input"
                  placeholder="Administrator"
                  :disabled="formLoading"
                />
              </div>
              <div class="form-row form-row--flex">
                <label class="form-label">RDP端口</label>
                <input
                  v-model="formData.rdp_port"
                  type="text"
                  class="form-input"
                  placeholder="3389"
                  :disabled="formLoading"
                />
              </div>
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

    <!-- Install Guide Dialog -->
    <Teleport to="body">
      <div v-if="showInstall" class="modal-overlay" @click.self="showInstall = false">
        <div class="modal-content" style="max-width: 640px;">
          <header class="modal-header">
            <h2 class="modal-title">安装监控插件</h2>
            <button class="modal-close" @click="showInstall = false">&#10005;</button>
          </header>

          <div v-if="installGuide" class="install-guide">
            <div class="install-info">
              <span class="install-asset">资产: <strong>{{ installGuide.asset_name }}</strong> (ID: {{ installGuide.asset_id }})</span>
              <span class="install-os">系统: <strong>{{ installGuide.os_type === 'windows' ? 'Windows' : 'Linux' }}</strong></span>
            </div>

            <div class="install-section">
              <h3 class="install-section-title">一键安装命令</h3>
              <div class="code-block">
                <code>{{ installGuide.commands[installGuide.os_type] || installGuide.commands['linux'] }}</code>
                <button class="copy-btn" @click="copyCommand(installGuide.os_type)">复制</button>
              </div>
            </div>

            <div class="install-section">
              <h3 class="install-section-title">手动安装</h3>
              <div class="code-block code-block--manual">
                <pre>{{ installGuide.commands[installGuide.os_type + '_manual'] || installGuide.commands['linux_manual'] }}</pre>
                <button class="copy-btn" @click="copyCommand(installGuide.os_type + '_manual')">复制</button>
              </div>
            </div>

            <p class="install-tip">请在目标服务器上以管理员权限执行以上命令</p>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- SSH Terminal Dialog -->
    <Teleport to="body">
      <div v-if="showTerminal" class="terminal-overlay">
        <div class="terminal-dialog">
          <header class="terminal-header">
            <div class="terminal-header-left">
              <span class="terminal-icon">⌨</span>
              <span class="terminal-title">
                SSH: {{ terminalAsset?.name }} ({{ terminalAsset?.ip }})
              </span>
              <span v-if="termConnecting" class="terminal-status terminal-status--connecting">连接中...</span>
              <span v-else-if="termConnected" class="terminal-status terminal-status--connected">已连接</span>
              <span v-else class="terminal-status terminal-status--disconnected">未连接</span>
            </div>
            <button class="terminal-close" @click="closeTerminal">✕</button>
          </header>

          <!-- Credential form (shown when no stored credentials) -->
          <div v-if="showCredForm" class="cred-form">
            <div class="cred-form-grid">
              <div class="form-row">
                <label class="form-label">主机</label>
                <input v-model="credForm.host" type="text" class="form-input" placeholder="10.0.1.100" />
              </div>
              <div class="form-row">
                <label class="form-label">端口</label>
                <input v-model="credForm.port" type="text" class="form-input" placeholder="22" />
              </div>
              <div class="form-row">
                <label class="form-label">用户名</label>
                <input v-model="credForm.user" type="text" class="form-input" placeholder="root" />
              </div>
              <div class="form-row">
                <label class="form-label">密码</label>
                <input v-model="credForm.password" type="password" class="form-input" placeholder="输入密码" />
              </div>
            </div>
            <button class="btn-submit" @click="connectSSH">连接</button>
          </div>

          <!-- Terminal container -->
          <div ref="terminalEl" class="terminal-container"></div>
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

.asset-id {
  font-family: var(--font-display);
  font-size: 0.75rem;
  color: var(--accent-amber);
  background: rgba(245, 158, 11, 0.08);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
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

.action-btns {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.btn-remote {
  padding: var(--space-1) var(--space-2);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  color: var(--accent-cyan);
  cursor: pointer;
  font-size: 0.75rem;
  transition: all var(--transition-fast);
  line-height: 1;
}

.btn-remote:hover {
  background: rgba(0, 212, 255, 0.08);
  border-color: rgba(0, 212, 255, 0.2);
  box-shadow: 0 0 8px rgba(0, 212, 255, 0.1);
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
  max-height: 70vh;
  overflow-y: auto;
}

.form-row {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-row--flex {
  flex: 1;
  min-width: 0;
}

.form-row-inline {
  display: flex;
  gap: var(--space-4);
}

.form-section-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--accent-teal);
  letter-spacing: 0.04em;
  text-transform: uppercase;
  padding-top: var(--space-2);
  border-top: 1px solid var(--color-border-subtle);
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

/* ====== Install Guide ====== */
.install-guide {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.install-info {
  display: flex;
  gap: var(--space-4);
  font-size: 0.82rem;
  color: var(--color-text-secondary);
}

.install-asset strong, .install-os strong {
  color: var(--accent-cyan);
}

.install-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.install-section-title {
  font-size: 0.82rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.code-block {
  position: relative;
  background: var(--color-bg-deep);
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  padding-right: 60px;
  font-family: var(--font-display);
  font-size: 0.72rem;
  color: var(--accent-emerald);
  line-height: 1.6;
  overflow-x: auto;
  word-break: break-all;
}

.code-block--manual {
  max-height: 260px;
  overflow-y: auto;
}

.code-block pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.copy-btn {
  position: absolute;
  top: var(--space-2);
  right: var(--space-2);
  padding: 2px 8px;
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid rgba(0, 212, 255, 0.3);
  border-radius: var(--radius-sm);
  color: var(--accent-cyan);
  font-size: 0.65rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.copy-btn:hover {
  background: rgba(0, 212, 255, 0.2);
  border-color: var(--accent-cyan);
}

.install-tip {
  font-size: 0.72rem;
  color: var(--accent-amber);
  margin: 0;
  padding: var(--space-2) var(--space-3);
  background: rgba(240, 160, 48, 0.08);
  border-radius: var(--radius-sm);
}

/* ====== SSH Terminal ====== */
.terminal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(6, 10, 18, 0.92);
  backdrop-filter: blur(6px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 300;
  padding: var(--space-6);
}

.terminal-dialog {
  width: 100%;
  max-width: 960px;
  height: 80vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-deep);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: 0 0 40px rgba(0, 212, 255, 0.08), 0 16px 60px rgba(0, 0, 0, 0.6);
  overflow: hidden;
}

.terminal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-5);
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.terminal-header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.terminal-icon {
  color: var(--accent-cyan);
  font-size: 0.9rem;
  filter: drop-shadow(0 0 4px rgba(0, 212, 255, 0.4));
}

.terminal-title {
  font-family: var(--font-display);
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-primary);
  letter-spacing: 0.02em;
}

.terminal-status {
  font-size: 0.7rem;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.terminal-status--connecting {
  color: var(--accent-amber);
  background: rgba(240, 160, 48, 0.12);
  animation: status-pulse 1.5s ease-in-out infinite;
}

.terminal-status--connected {
  color: var(--accent-emerald);
  background: rgba(45, 212, 160, 0.1);
}

.terminal-status--disconnected {
  color: var(--color-text-muted);
  background: rgba(77, 98, 130, 0.15);
}

@keyframes status-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.terminal-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  font-size: 0.85rem;
  transition: all var(--transition-fast);
}

.terminal-close:hover {
  background: rgba(240, 72, 72, 0.08);
  border-color: rgba(240, 72, 72, 0.2);
  color: var(--accent-red);
}

.terminal-container {
  flex: 1;
  padding: var(--space-2);
  background: var(--color-bg-deep);
  overflow: hidden;
}

.terminal-container :deep(.xterm) {
  height: 100%;
}

.terminal-container :deep(.xterm-viewport) {
  overflow-y: auto !important;
}

/* Credential form inside terminal */
.cred-form {
  padding: var(--space-6);
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  align-items: center;
  justify-content: center;
  flex: 1;
}

.cred-form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
  width: 100%;
  max-width: 420px;
}

.cred-form .btn-submit {
  margin-top: var(--space-2);
}
</style>
