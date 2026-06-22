<script setup lang="ts">
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>

<template>
  <div class="dashboard">
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
          <span class="info-value">{{ userStore.roleId === 1 ? '管理员' : userStore.roleId === 2 ? '普通用户' : '未分配' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  max-width: 960px;
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
