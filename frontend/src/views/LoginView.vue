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
    <div class="login-bg-grid"></div>
    <div class="login-container">
      <div class="login-brand">
        <div class="brand-icon">⚡</div>
        <h1 class="brand-title">ITSM Ops</h1>
        <p class="brand-subtitle">智能运维监控平台</p>
      </div>

      <form class="login-form" @submit.prevent="handleLogin">
        <div class="form-group">
          <label class="form-label" for="username">用户名</label>
          <div class="input-wrapper">
            <span class="input-icon">⬡</span>
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
        </div>

        <div class="form-group">
          <label class="form-label" for="password">密码</label>
          <div class="input-wrapper">
            <span class="input-icon">◈</span>
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
        </div>

        <div v-if="errorMsg" class="form-error">
          <span class="error-icon">▲</span>
          {{ errorMsg }}
        </div>

        <button type="submit" class="login-btn" :disabled="loading">
          <span v-if="loading" class="btn-spinner"></span>
          <span v-else>登 录</span>
        </button>
      </form>

      <footer class="login-footer">
        <span>ITSM Ops v0.1.0</span>
      </footer>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background:
    radial-gradient(ellipse 80% 50% at 50% -20%, rgba(0, 212, 255, 0.08) 0%, transparent 60%),
    radial-gradient(ellipse 60% 40% at 80% 100%, rgba(46, 165, 160, 0.05) 0%, transparent 50%),
    var(--color-bg-deep);
  position: relative;
  overflow: hidden;
}

.login-bg-grid {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(0, 212, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(0, 212, 255, 0.03) 1px, transparent 1px);
  background-size: 60px 60px;
  mask-image: radial-gradient(ellipse 60% 60% at 50% 50%, black 20%, transparent 70%);
  pointer-events: none;
}

.login-container {
  width: 100%;
  max-width: 400px;
  padding: var(--space-10);
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
  position: relative;
  z-index: 1;
}

.login-brand {
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-3);
}

.brand-icon {
  font-size: 2.5rem;
  line-height: 1;
  filter: drop-shadow(0 0 16px rgba(0, 212, 255, 0.5));
  animation: brand-glow 3s ease-in-out infinite;
}

@keyframes brand-glow {
  0%, 100% { filter: drop-shadow(0 0 16px rgba(0, 212, 255, 0.5)); }
  50% { filter: drop-shadow(0 0 28px rgba(0, 212, 255, 0.8)); }
}

.brand-title {
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: var(--color-text-primary);
}

.brand-subtitle {
  font-size: 0.85rem;
  color: var(--color-text-muted);
  letter-spacing: 0.06em;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-8) var(--space-6);
  box-shadow: var(--shadow-card), var(--shadow-glow-cyan);
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

.form-label {
  font-size: 0.8rem;
  color: var(--color-text-secondary);
  letter-spacing: 0.04em;
  font-weight: 500;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  background: var(--color-bg-deep);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-3) var(--space-4);
  transition: border-color var(--transition-fast), box-shadow var(--transition-fast);
}

.input-wrapper:focus-within {
  border-color: var(--accent-cyan);
  box-shadow: 0 0 0 2px rgba(0, 212, 255, 0.15);
}

.input-icon {
  font-size: 0.9rem;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.form-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--color-text-primary);
  font-family: var(--font-body);
  font-size: 0.9rem;
  line-height: 1.5;
}

.form-input::placeholder {
  color: var(--color-text-muted);
}

.form-input:disabled {
  opacity: 0.6;
  cursor: not-allowed;
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

.error-icon {
  font-size: 0.7rem;
  flex-shrink: 0;
}

.login-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 44px;
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.15) 0%, rgba(0, 212, 255, 0.05) 100%);
  border: 1px solid var(--accent-cyan);
  border-radius: var(--radius-md);
  color: var(--accent-cyan);
  font-family: var(--font-display);
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.15em;
  cursor: pointer;
  transition: all var(--transition-fast);
  text-transform: uppercase;
}

.login-btn:hover:not(:disabled) {
  background: linear-gradient(135deg, rgba(0, 212, 255, 0.25) 0%, rgba(0, 212, 255, 0.1) 100%);
  box-shadow: 0 0 20px rgba(0, 212, 255, 0.2);
}

.login-btn:active:not(:disabled) {
  transform: scale(0.98);
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid rgba(0, 212, 255, 0.3);
  border-top-color: var(--accent-cyan);
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-footer {
  text-align: center;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

@media (max-width: 480px) {
  .login-container {
    padding: var(--space-6) var(--space-4);
  }

  .login-form {
    padding: var(--space-6) var(--space-4);
  }

  .brand-title {
    font-size: 1.5rem;
  }
}
</style>
