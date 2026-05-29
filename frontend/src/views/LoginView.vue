<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/store/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMsg = ref('')

async function handleLogin(): Promise<void> {
  if (!username.value || !password.value) {
    errorMsg.value = '请输入用户名和密码'
    return
  }

  loading.value = true
  errorMsg.value = ''

  try {
    await auth.login(username.value, password.value)
    const redirect = (route.query.redirect as string) || '/dashboard'
    await router.push(redirect)
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMsg.value = err.message || '登录失败，请检查用户名和密码'
    } else {
      errorMsg.value = '登录失败，请检查用户名和密码'
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-left">
      <div class="login-illustration">
        <div class="brand-logo">📋</div>
        <h1 class="brand-title">ITSM Ops</h1>
        <p class="brand-subtitle">智能运维监控平台</p>
        <div class="brand-features">
          <div class="feature-item">
            <span class="feature-dot"></span>
            <span>实时监控</span>
          </div>
          <div class="feature-item">
            <span class="feature-dot"></span>
            <span>智能告警</span>
          </div>
          <div class="feature-item">
            <span class="feature-dot"></span>
            <span>远程管理</span>
          </div>
        </div>
      </div>
    </div>

    <div class="login-right">
      <div class="login-card">
        <h2 class="login-title">欢迎回来</h2>
        <p class="login-desc">登录以继续使用运维监控平台</p>

        <form class="login-form" @submit.prevent="handleLogin">
          <div v-if="errorMsg" class="form-error">{{ errorMsg }}</div>

          <div class="form-group">
            <label class="form-label" for="username">用户名</label>
            <input
              id="username"
              v-model="username"
              type="text"
              class="form-input"
              placeholder="请输入用户名"
              autocomplete="username"
              :disabled="loading"
            />
          </div>

          <div class="form-group">
            <label class="form-label" for="password">密码</label>
            <input
              id="password"
              v-model="password"
              type="password"
              class="form-input"
              placeholder="请输入密码"
              autocomplete="current-password"
              :disabled="loading"
            />
          </div>

          <button type="submit" class="login-btn" :disabled="loading">
            <span v-if="loading" class="btn-loading"></span>
            <span v-else>登 录</span>
          </button>
        </form>

        <p class="login-footer">ITSM Ops © 2026</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  min-height: 100vh;
  background: var(--color-bg-base);
}

.login-left {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #FF2442 0%, #FF6B81 50%, #FFB3C1 100%);
  position: relative;
  overflow: hidden;
}

.login-left::before {
  content: '';
  position: absolute;
  width: 400px;
  height: 400px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  top: -100px;
  right: -100px;
}

.login-left::after {
  content: '';
  position: absolute;
  width: 200px;
  height: 200px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.08);
  bottom: -50px;
  left: -50px;
}

.login-illustration {
  text-align: center;
  color: white;
  position: relative;
  z-index: 1;
}

.brand-logo {
  font-size: 4rem;
  margin-bottom: var(--space-5);
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.15));
}

.brand-title {
  font-size: 2.5rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  margin-bottom: var(--space-3);
}

.brand-subtitle {
  font-size: 1.1rem;
  opacity: 0.9;
  margin-bottom: var(--space-10);
}

.brand-features {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  align-items: center;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: 0.95rem;
  opacity: 0.9;
}

.feature-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: white;
}

.login-right {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-8);
  background: var(--color-bg-base);
}

.login-card {
  width: 100%;
  max-width: 400px;
}

.login-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin-bottom: var(--space-2);
}

.login-desc {
  font-size: 0.9rem;
  color: var(--color-text-muted);
  margin-bottom: var(--space-8);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

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
  gap: var(--space-2);
}

.form-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.form-input {
  padding: var(--space-3) var(--space-4);
  border: 1.5px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  font-size: 0.95rem;
  color: var(--color-text-primary);
  background: var(--color-bg-base);
  transition: all var(--transition-fast);
  outline: none;
  font-family: var(--font-body);
}

.form-input::placeholder {
  color: var(--color-text-muted);
}

.form-input:focus {
  border-color: var(--accent-red);
  box-shadow: 0 0 0 3px rgba(255, 36, 66, 0.1);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-btn {
  padding: var(--space-3) var(--space-6);
  border: none;
  border-radius: var(--radius-md);
  background: var(--accent-red);
  color: white;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-body);
  margin-top: var(--space-2);
}

.login-btn:hover:not(:disabled) {
  background: #E6203C;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(255, 36, 66, 0.3);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.btn-loading {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-footer {
  margin-top: var(--space-10);
  text-align: center;
  font-size: 0.8rem;
  color: var(--color-text-muted);
}

@media (max-width: 768px) {
  .login-left {
    display: none;
  }

  .login-right {
    padding: var(--space-6);
  }
}
</style>
