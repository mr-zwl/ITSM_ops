<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { post } from '@/api/http'

const auth = useAuthStore()
const router = useRouter()
const sidebarCollapsed = ref(false)
const userMenuOpen = ref(false)

const showChangePwd = ref(false)
const changePwdForm = ref({ old_password: '', new_password: '', confirm_password: '' })
const changePwdLoading = ref(false)
const changePwdMsg = ref('')

const showRegister = ref(false)
const registerForm = ref({ username: '', password: '', confirm_password: '', display_name: '', email: '', phone: '' })
const registerLoading = ref(false)
const registerMsg = ref('')

function toggleSidebar(): void {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

async function handleLogout(): Promise<void> {
  userMenuOpen.value = false
  auth.logout()
  await router.push('/login')
}

function toggleUserMenu(): void {
  userMenuOpen.value = !userMenuOpen.value
}

function openChangePwd(): void {
  userMenuOpen.value = false
  changePwdForm.value = { old_password: '', new_password: '', confirm_password: '' }
  changePwdMsg.value = ''
  showChangePwd.value = true
}

function openRegister(): void {
  userMenuOpen.value = false
  registerForm.value = { username: '', password: '', confirm_password: '', display_name: '', email: '', phone: '' }
  registerMsg.value = ''
  showRegister.value = true
}

async function submitChangePwd(): Promise<void> {
  const f = changePwdForm.value
  if (!f.old_password || !f.new_password || !f.confirm_password) {
    changePwdMsg.value = '请填写所有字段'
    return
  }
  if (f.new_password !== f.confirm_password) {
    changePwdMsg.value = '两次输入的新密码不一致'
    return
  }
  if (f.new_password.length < 6) {
    changePwdMsg.value = '新密码至少6位'
    return
  }
  changePwdLoading.value = true
  changePwdMsg.value = ''
  try {
    await post('/auth/change-password', { old_password: f.old_password, new_password: f.new_password })
    changePwdMsg.value = '密码修改成功'
    setTimeout(() => { showChangePwd.value = false }, 1500)
  } catch (e: any) {
    changePwdMsg.value = e?.message || '修改失败'
  } finally {
    changePwdLoading.value = false
  }
}

async function submitRegister(): Promise<void> {
  const f = registerForm.value
  if (!f.username || !f.password || !f.confirm_password) {
    registerMsg.value = '请填写用户名和密码'
    return
  }
  if (f.password !== f.confirm_password) {
    registerMsg.value = '两次输入的密码不一致'
    return
  }
  if (f.username.length < 3) {
    registerMsg.value = '用户名至少3位'
    return
  }
  if (f.password.length < 6) {
    registerMsg.value = '密码至少6位'
    return
  }
  registerLoading.value = true
  registerMsg.value = ''
  try {
    await post('/auth/register', {
      username: f.username,
      password: f.password,
      display_name: f.display_name || f.username,
      email: f.email,
      phone: f.phone,
    })
    registerMsg.value = '用户创建成功'
    setTimeout(() => { showRegister.value = false }, 1500)
  } catch (e: any) {
    registerMsg.value = e?.message || '注册失败'
  } finally {
    registerLoading.value = false
  }
}

function handleClickOutside(e: MouseEvent): void {
  const target = e.target as HTMLElement
  if (!target.closest('.user-menu-wrapper')) {
    userMenuOpen.value = false
  }
}

const currentTime = ref('')
let timer: ReturnType<typeof setInterval> | undefined

function formatTime(): string {
  const now = new Date()
  return now.toLocaleString('zh-CN', {
    year: 'numeric', month: '2-digit', day: '2-digit',
    hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false,
  })
}

onMounted(() => {
  currentTime.value = formatTime()
  timer = setInterval(() => { currentTime.value = formatTime() }, 1000)
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="main-layout">
    <aside class="sidebar" :class="{ 'sidebar--collapsed': sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo-mark">📋</div>
        <transition name="fade">
          <div v-if="!sidebarCollapsed" class="sidebar-brand">
            <span class="brand-title">ITSM Ops</span>
          </div>
        </transition>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/dashboard" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">📊</span>
          <transition name="fade"><span v-if="!sidebarCollapsed" class="nav-label">仪表盘</span></transition>
        </router-link>
        <router-link to="/assets" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">💻</span>
          <transition name="fade"><span v-if="!sidebarCollapsed" class="nav-label">资产管理</span></transition>
        </router-link>
        <router-link to="/metrics" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">📈</span>
          <transition name="fade"><span v-if="!sidebarCollapsed" class="nav-label">指标监控</span></transition>
        </router-link>
        <router-link to="/alerts" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">🔔</span>
          <transition name="fade"><span v-if="!sidebarCollapsed" class="nav-label">告警事件</span></transition>
        </router-link>
      </nav>
    </aside>

    <div class="main-content">
      <header class="top-bar">
        <button class="collapse-btn" @click="toggleSidebar">
          <span class="collapse-icon" :class="{ 'collapse-icon--flipped': sidebarCollapsed }">‹</span>
        </button>
        <div class="top-bar-spacer"></div>
        <div class="top-bar-info">
          <time class="top-time">{{ currentTime }}</time>
          <div class="user-menu-wrapper">
            <button class="user-badge" @click.stop="toggleUserMenu">
              <span class="user-avatar">{{ auth.user?.username?.charAt(0)?.toUpperCase() || '?' }}</span>
              <span class="user-name">{{ auth.user?.username || '未知用户' }}</span>
              <span class="user-arrow" :class="{ 'user-arrow--open': userMenuOpen }">▾</span>
            </button>
            <transition name="dropdown">
              <div v-if="userMenuOpen" class="user-dropdown">
                <div class="dropdown-header">
                  <span class="dropdown-avatar">{{ auth.user?.username?.charAt(0)?.toUpperCase() || '?' }}</span>
                  <div class="dropdown-user-info">
                    <div class="dropdown-username">{{ auth.user?.username }}</div>
                    <div class="dropdown-role">管理员</div>
                  </div>
                </div>
                <div class="dropdown-divider"></div>
                <button class="dropdown-item" @click="openChangePwd">
                  <span class="dropdown-icon">🔑</span>修改密码
                </button>
                <button class="dropdown-item" @click="openRegister">
                  <span class="dropdown-icon">👤</span>注册用户
                </button>
                <div class="dropdown-divider"></div>
                <button class="dropdown-item dropdown-item--danger" @click="handleLogout">
                  <span class="dropdown-icon">🚪</span>退出登录
                </button>
              </div>
            </transition>
          </div>
        </div>
      </header>

      <main class="page-content">
        <router-view />
      </main>
    </div>

    <!-- Change Password Dialog -->
    <Teleport to="body">
      <div v-if="showChangePwd" class="dialog-overlay" @click.self="showChangePwd = false">
        <div class="dialog">
          <div class="dialog-header">
            <h3>修改密码</h3>
            <button class="dialog-close" @click="showChangePwd = false">✕</button>
          </div>
          <div class="dialog-body">
            <label class="form-label">旧密码</label>
            <input v-model="changePwdForm.old_password" type="password" class="form-input" placeholder="请输入旧密码" />
            <label class="form-label">新密码</label>
            <input v-model="changePwdForm.new_password" type="password" class="form-input" placeholder="请输入新密码（至少6位）" />
            <label class="form-label">确认新密码</label>
            <input v-model="changePwdForm.confirm_password" type="password" class="form-input" placeholder="请再次输入新密码" />
            <div v-if="changePwdMsg" class="form-msg" :class="{ 'form-msg--success': changePwdMsg.includes('成功') }">{{ changePwdMsg }}</div>
          </div>
          <div class="dialog-footer">
            <button class="btn btn-secondary" @click="showChangePwd = false">取消</button>
            <button class="btn btn-primary" @click="submitChangePwd" :disabled="changePwdLoading">
              {{ changePwdLoading ? '提交中...' : '确认修改' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Register User Dialog -->
    <Teleport to="body">
      <div v-if="showRegister" class="dialog-overlay" @click.self="showRegister = false">
        <div class="dialog">
          <div class="dialog-header">
            <h3>注册新用户</h3>
            <button class="dialog-close" @click="showRegister = false">✕</button>
          </div>
          <div class="dialog-body">
            <label class="form-label">用户名 *</label>
            <input v-model="registerForm.username" type="text" class="form-input" placeholder="请输入用户名（至少3位）" />
            <label class="form-label">密码 *</label>
            <input v-model="registerForm.password" type="password" class="form-input" placeholder="请输入密码（至少6位）" />
            <label class="form-label">确认密码 *</label>
            <input v-model="registerForm.confirm_password" type="password" class="form-input" placeholder="请再次输入密码" />
            <label class="form-label">显示名称</label>
            <input v-model="registerForm.display_name" type="text" class="form-input" placeholder="请输入显示名称" />
            <label class="form-label">邮箱</label>
            <input v-model="registerForm.email" type="email" class="form-input" placeholder="请输入邮箱" />
            <label class="form-label">手机</label>
            <input v-model="registerForm.phone" type="tel" class="form-input" placeholder="请输入手机号" />
            <div v-if="registerMsg" class="form-msg" :class="{ 'form-msg--success': registerMsg.includes('成功') }">{{ registerMsg }}</div>
          </div>
          <div class="dialog-footer">
            <button class="btn btn-secondary" @click="showRegister = false">取消</button>
            <button class="btn btn-primary" @click="submitRegister" :disabled="registerLoading">
              {{ registerLoading ? '提交中...' : '创建用户' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.main-layout { display: flex; min-height: 100vh; background: var(--color-bg-deep); }
.sidebar { width: 220px; display: flex; flex-direction: column; background: var(--color-bg-base); border-right: 1px solid var(--color-border); transition: width var(--transition-base); flex-shrink: 0; position: relative; z-index: 50; }
.sidebar--collapsed { width: 64px; }
.sidebar-header { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-5) var(--space-4); border-bottom: 1px solid var(--color-border); min-height: 64px; }
.logo-mark { font-size: 1.5rem; line-height: 1; flex-shrink: 0; }
.sidebar-brand { overflow: hidden; white-space: nowrap; }
.brand-title { font-size: 1.1rem; font-weight: 700; letter-spacing: 0.02em; color: var(--accent-red); }
.sidebar-nav { flex: 1; display: flex; flex-direction: column; gap: var(--space-1); padding: var(--space-3); }
.nav-item { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-md); color: var(--color-text-secondary); text-decoration: none; font-size: 0.9rem; font-weight: 500; transition: all var(--transition-fast); cursor: pointer; border: none; background: none; width: 100%; font-family: var(--font-body); white-space: nowrap; }
.nav-item:hover { background: var(--color-bg-elevated); color: var(--color-text-primary); }
.nav-item--active { background: var(--accent-pink); color: var(--accent-red); font-weight: 600; }
.nav-icon { font-size: 1.15rem; flex-shrink: 0; width: 24px; text-align: center; }
.nav-label { overflow: hidden; }
.fade-enter-active, .fade-leave-active { transition: opacity var(--transition-fast); }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.main-content { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.top-bar { display: flex; align-items: center; padding: var(--space-3) var(--space-6); border-bottom: 1px solid var(--color-border); background: rgba(255,255,255,0.9); backdrop-filter: blur(12px); position: sticky; top: 0; z-index: 40; min-height: 56px; }
.collapse-btn { display: flex; align-items: center; justify-content: center; width: 36px; height: 36px; border-radius: var(--radius-md); border: 1px solid var(--color-border); background: var(--color-bg-base); color: var(--color-text-muted); cursor: pointer; transition: all var(--transition-fast); font-size: 1.1rem; font-weight: 300; }
.collapse-btn:hover { border-color: var(--accent-red); color: var(--accent-red); }
.collapse-icon { display: inline-block; transition: transform var(--transition-base); }
.collapse-icon--flipped { transform: rotate(180deg); }
.top-bar-spacer { flex: 1; }
.top-bar-info { display: flex; align-items: center; gap: var(--space-5); }
.top-time { font-size: 0.85rem; color: var(--color-text-muted); letter-spacing: 0.02em; }
.user-menu-wrapper { position: relative; }
.user-badge { display: flex; align-items: center; gap: var(--space-2); cursor: pointer; padding: var(--space-1) var(--space-2); border-radius: var(--radius-lg); border: none; background: none; transition: background var(--transition-fast); }
.user-badge:hover { background: var(--color-bg-elevated); }
.user-avatar { width: 32px; height: 32px; border-radius: 50%; background: var(--accent-pink); display: flex; align-items: center; justify-content: center; font-size: 0.8rem; font-weight: 700; color: var(--accent-red); }
.user-name { font-size: 0.85rem; color: var(--color-text-secondary); font-weight: 500; }
.user-arrow { font-size: 0.7rem; color: var(--color-text-muted); transition: transform var(--transition-fast); }
.user-arrow--open { transform: rotate(180deg); }
.user-dropdown { position: absolute; top: calc(100% + 8px); right: 0; width: 200px; background: var(--color-bg-base); border: 1px solid var(--color-border); border-radius: var(--radius-lg); box-shadow: 0 8px 24px rgba(0,0,0,0.12); z-index: 100; overflow: hidden; }
.dropdown-enter-active, .dropdown-leave-active { transition: all 0.15s ease; }
.dropdown-enter-from, .dropdown-leave-to { opacity: 0; transform: translateY(-8px); }
.dropdown-header { display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); }
.dropdown-avatar { width: 36px; height: 36px; border-radius: 50%; background: var(--accent-pink); display: flex; align-items: center; justify-content: center; font-size: 0.85rem; font-weight: 700; color: var(--accent-red); flex-shrink: 0; }
.dropdown-user-info { min-width: 0; }
.dropdown-username { font-size: 0.9rem; font-weight: 600; color: var(--color-text-primary); }
.dropdown-role { font-size: 0.75rem; color: var(--color-text-muted); }
.dropdown-divider { height: 1px; background: var(--color-border); }
.dropdown-item { display: flex; align-items: center; gap: var(--space-3); width: 100%; padding: var(--space-3) var(--space-4); border: none; background: none; color: var(--color-text-secondary); font-size: 0.85rem; cursor: pointer; transition: all var(--transition-fast); font-family: var(--font-body); text-align: left; }
.dropdown-item:hover { background: var(--color-bg-elevated); color: var(--color-text-primary); }
.dropdown-item--danger { color: var(--accent-red); }
.dropdown-item--danger:hover { background: var(--accent-pink); color: var(--accent-red); }
.dropdown-icon { font-size: 1rem; width: 20px; text-align: center; }
.dialog-overlay { position: fixed; inset: 0; background: rgba(0,0,0,0.4); display: flex; align-items: center; justify-content: center; z-index: 200; }
.dialog { background: var(--color-bg-base); border-radius: var(--radius-xl); box-shadow: 0 16px 48px rgba(0,0,0,0.16); width: 420px; max-width: 90vw; max-height: 85vh; overflow-y: auto; }
.dialog-header { display: flex; align-items: center; justify-content: space-between; padding: var(--space-4) var(--space-5); border-bottom: 1px solid var(--color-border); }
.dialog-header h3 { font-size: 1.05rem; font-weight: 600; color: var(--color-text-primary); margin: 0; }
.dialog-close { display: flex; align-items: center; justify-content: center; width: 28px; height: 28px; border-radius: var(--radius-md); border: none; background: none; color: var(--color-text-muted); cursor: pointer; font-size: 0.9rem; transition: all var(--transition-fast); }
.dialog-close:hover { background: var(--color-bg-elevated); color: var(--color-text-primary); }
.dialog-body { padding: var(--space-5); display: flex; flex-direction: column; gap: var(--space-3); }
.form-label { font-size: 0.82rem; font-weight: 500; color: var(--color-text-secondary); margin-bottom: -4px; }
.form-input { width: 100%; padding: var(--space-2) var(--space-3); border: 1px solid var(--color-border); border-radius: var(--radius-md); font-size: 0.9rem; color: var(--color-text-primary); background: var(--color-bg-base); transition: border-color var(--transition-fast); outline: none; font-family: var(--font-body); box-sizing: border-box; }
.form-input:focus { border-color: var(--accent-red); box-shadow: 0 0 0 3px var(--accent-pink); }
.form-msg { font-size: 0.82rem; color: var(--accent-red); padding: var(--space-1) 0; }
.form-msg--success { color: #52c41a; }
.dialog-footer { display: flex; justify-content: flex-end; gap: var(--space-3); padding: var(--space-3) var(--space-5) var(--space-5); }
.btn { padding: var(--space-2) var(--space-4); border-radius: var(--radius-md); font-size: 0.85rem; font-weight: 500; cursor: pointer; border: none; transition: all var(--transition-fast); font-family: var(--font-body); }
.btn:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-secondary { background: var(--color-bg-elevated); color: var(--color-text-secondary); }
.btn-secondary:hover { background: var(--color-border); }
.btn-primary { background: var(--accent-red); color: white; }
.btn-primary:hover:not(:disabled) { background: #e01f3a; }
.page-content { flex: 1; padding: var(--space-6); max-width: 1440px; width: 100%; }
@media (max-width: 768px) {
  .sidebar { position: fixed; left: 0; top: 0; bottom: 0; width: 220px; transform: translateX(-100%); transition: transform var(--transition-base); z-index: 100; }
  .sidebar:not(.sidebar--collapsed) { transform: translateX(0); }
  .sidebar--collapsed { transform: translateX(-100%); width: 220px; }
  .top-bar { padding: var(--space-3) var(--space-4); }
  .page-content { padding: var(--space-4); }
  .top-time { display: none; }
}
</style>
