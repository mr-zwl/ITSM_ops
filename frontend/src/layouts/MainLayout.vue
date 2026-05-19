<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const auth = useAuthStore()
const router = useRouter()
const sidebarCollapsed = ref(false)

function toggleSidebar(): void {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

async function handleLogout(): Promise<void> {
  auth.logout()
  await router.push('/login')
}

const currentTime = ref('')
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

onMounted(() => {
  currentTime.value = formatTime()
  timer = setInterval(() => {
    currentTime.value = formatTime()
  }, 1000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <div class="main-layout">
    <aside class="sidebar" :class="{ 'sidebar--collapsed': sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo-mark">⚡</div>
        <transition name="fade">
          <div v-if="!sidebarCollapsed" class="sidebar-brand">
            <span class="brand-title">ITSM Ops</span>
          </div>
        </transition>
      </div>

      <nav class="sidebar-nav">
        <router-link to="/dashboard" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">◉</span>
          <transition name="fade">
            <span v-if="!sidebarCollapsed" class="nav-label">仪表盘</span>
          </transition>
        </router-link>

        <router-link to="/assets" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">⬡</span>
          <transition name="fade">
            <span v-if="!sidebarCollapsed" class="nav-label">资产管理</span>
          </transition>
        </router-link>

        <router-link to="/metrics" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">◈</span>
          <transition name="fade">
            <span v-if="!sidebarCollapsed" class="nav-label">指标监控</span>
          </transition>
        </router-link>

        <router-link to="/alerts" class="nav-item" active-class="nav-item--active">
          <span class="nav-icon">▲</span>
          <transition name="fade">
            <span v-if="!sidebarCollapsed" class="nav-label">告警事件</span>
          </transition>
        </router-link>
      </nav>

      <div class="sidebar-footer">
        <button class="nav-item nav-item--logout" @click="handleLogout">
          <span class="nav-icon">⏻</span>
          <transition name="fade">
            <span v-if="!sidebarCollapsed" class="nav-label">退出登录</span>
          </transition>
        </button>
      </div>
    </aside>

    <div class="main-content">
      <header class="top-bar">
        <button class="collapse-btn" @click="toggleSidebar" :title="sidebarCollapsed ? '展开侧栏' : '收起侧栏'">
          <span class="collapse-icon" :class="{ 'collapse-icon--flipped': sidebarCollapsed }">◅</span>
        </button>

        <div class="top-bar-spacer"></div>

        <div class="top-bar-info">
          <time class="top-time">{{ currentTime }}</time>
          <div class="user-badge">
            <span class="user-avatar">{{ auth.user?.username?.charAt(0)?.toUpperCase() || '?' }}</span>
            <span class="user-name">{{ auth.user?.username || '未知用户' }}</span>
          </div>
        </div>
      </header>

      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<style scoped>
.main-layout {
  display: flex;
  min-height: 100vh;
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 212, 255, 0.06) 0%, transparent 60%),
    radial-gradient(ellipse 60% 40% at 80% 100%, rgba(46, 165, 160, 0.04) 0%, transparent 50%),
    var(--color-bg-deep);
}

/* ====== Sidebar ====== */
.sidebar {
  width: 220px;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-base);
  border-right: 1px solid var(--color-border);
  transition: width var(--transition-base);
  flex-shrink: 0;
  position: relative;
  z-index: 50;
}

.sidebar--collapsed {
  width: 60px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-5) var(--space-4);
  border-bottom: 1px solid var(--color-border);
  min-height: 56px;
}

.logo-mark {
  font-size: 1.3rem;
  line-height: 1;
  filter: drop-shadow(0 0 8px rgba(0, 212, 255, 0.5));
  flex-shrink: 0;
}

.sidebar-brand {
  overflow: hidden;
  white-space: nowrap;
}

.brand-title {
  font-family: var(--font-display);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--color-text-primary);
}

.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  padding: var(--space-4) var(--space-2);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3) var(--space-3);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all var(--transition-fast);
  cursor: pointer;
  border: none;
  background: none;
  width: 100%;
  font-family: var(--font-body);
  white-space: nowrap;
}

.nav-item:hover {
  background: rgba(0, 212, 255, 0.06);
  color: var(--color-text-primary);
}

.nav-item--active {
  background: rgba(0, 212, 255, 0.08);
  color: var(--accent-cyan);
  border-left: 2px solid var(--accent-cyan);
}

.nav-item--logout {
  color: var(--color-text-muted);
}

.nav-item--logout:hover {
  color: var(--accent-red);
  background: rgba(240, 72, 72, 0.06);
}

.nav-icon {
  font-size: 1rem;
  flex-shrink: 0;
  width: 20px;
  text-align: center;
}

.nav-label {
  overflow: hidden;
}

.sidebar-footer {
  padding: var(--space-3) var(--space-2);
  border-top: 1px solid var(--color-border);
}

/* ====== Fade transition ====== */
.fade-enter-active,
.fade-leave-active {
  transition: opacity var(--transition-fast);
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* ====== Top Bar ====== */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.top-bar {
  display: flex;
  align-items: center;
  padding: var(--space-3) var(--space-6);
  border-bottom: 1px solid var(--color-border);
  background: linear-gradient(180deg, rgba(11, 17, 32, 0.95) 0%, rgba(11, 17, 32, 0.8) 100%);
  backdrop-filter: blur(12px);
  position: sticky;
  top: 0;
  z-index: 40;
  min-height: 52px;
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border);
  background: var(--color-bg-surface);
  color: var(--color-text-muted);
  cursor: pointer;
  transition: all var(--transition-fast);
  font-size: 0.7rem;
}

.collapse-btn:hover {
  border-color: var(--accent-cyan);
  color: var(--accent-cyan);
}

.collapse-icon {
  display: inline-block;
  transition: transform var(--transition-base);
}

.collapse-icon--flipped {
  transform: rotate(180deg);
}

.top-bar-spacer {
  flex: 1;
}

.top-bar-info {
  display: flex;
  align-items: center;
  gap: var(--space-5);
}

.top-time {
  font-family: var(--font-display);
  font-size: 0.8rem;
  color: var(--color-text-muted);
  letter-spacing: 0.04em;
}

.user-badge {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(0, 212, 255, 0.1);
  border: 1px solid rgba(0, 212, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-family: var(--font-display);
  font-size: 0.75rem;
  font-weight: 700;
  color: var(--accent-cyan);
}

.user-name {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

/* ====== Page Content ====== */
.page-content {
  flex: 1;
  padding: var(--space-8);
  max-width: 1440px;
  width: 100%;
}

/* ====== Responsive ====== */
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 220px;
    transform: translateX(-100%);
    transition: transform var(--transition-base);
    z-index: 100;
  }

  .sidebar:not(.sidebar--collapsed) {
    transform: translateX(0);
  }

  .sidebar--collapsed {
    transform: translateX(-100%);
    width: 220px;
  }

  .top-bar {
    padding: var(--space-3) var(--space-4);
  }

  .page-content {
    padding: var(--space-4);
  }

  .top-time {
    display: none;
  }
}
</style>
