<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useRouter } from 'vue-router'

const userStore = useUserStore()
const router = useRouter()

onMounted(() => {
  userStore.fetchProfile()
})

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <div class="dashboard">
    <header class="top-bar">
      <div class="top-bar-left">
        <span class="brand">IAM Platform</span>
      </div>
      <div class="top-bar-right">
        <span class="greeting">{{ userStore.username }}</span>
        <a-button type="text" size="small" @click="handleLogout">
          <template #icon>
            <icon-export />
          </template>
          退出
        </a-button>
      </div>
    </header>

    <main class="main-content">
      <div class="welcome-card">
        <h2>欢迎使用 IAM Platform</h2>
        <p>您已通过身份认证，可以访问受保护的资源。</p>
        <div class="info-grid">
          <div class="info-item">
            <span class="info-label">用户名</span>
            <span class="info-value">{{ userStore.username }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">角色</span>
            <span class="info-value">{{ userStore.roleId === 1 ? '管理员' : '普通用户' }}</span>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: #f5f7fa;
}

.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #e5e6eb;
}

.brand {
  font-size: 16px;
  font-weight: 600;
  color: #1d2129;
}

.top-bar-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.greeting {
  font-size: 14px;
  color: #4e5969;
}

.main-content {
  max-width: 960px;
  margin: 0 auto;
  padding: 32px 24px;
}

.welcome-card {
  background: #fff;
  border-radius: 8px;
  padding: 32px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.welcome-card h2 {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #1d2129;
}

.welcome-card > p {
  color: #86909c;
  margin-bottom: 24px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 16px;
  background: #f7f8fa;
  border-radius: 6px;
}

.info-label {
  font-size: 12px;
  color: #86909c;
}

.info-value {
  font-size: 14px;
  color: #1d2129;
  font-weight: 500;
}
</style>
