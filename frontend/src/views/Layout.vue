<script setup lang="ts">
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { useRouter, useRoute } from 'vue-router'
import { IconApps, IconExport } from '@arco-design/web-vue/es/icon'

const userStore = useUserStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  userStore.fetchProfile()
  userStore.fetchPermissions()
})

function handleMenuClick(key: string) {
  router.push(key)
}

function handleLogout() {
  userStore.logout()
}
</script>

<template>
  <a-layout class="app-layout">
    <!-- 顶栏 -->
    <a-layout-header class="top-bar">
      <div class="top-bar-left">
        <span class="brand">IAM Platform</span>
      </div>
      <div class="top-bar-right">
        <span class="greeting">{{ userStore.username }}</span>
        <a-button type="text" size="small" @click="handleLogout">
          <template #icon><icon-export /></template>
          退出
        </a-button>
      </div>
    </a-layout-header>

    <a-layout class="body-layout">
      <!-- 侧边栏 -->
      <a-layout-sider class="sidebar" :width="220" collapsible breakpoint="lg">
        <a-menu
          :selected-keys="[route.path]"
          @menu-item-click="handleMenuClick"
        >
          <a-menu-item key="/dashboard">
            <template #icon><icon-apps /></template>
            仪表盘
          </a-menu-item>
          <a-menu-item
            v-for="item in userStore.sidebarMenus"
            :key="'/' + item.path.replace('/system/', 'system/')"
          >
            {{ item.name }}
          </a-menu-item>
        </a-menu>
      </a-layout-sider>

      <!-- 内容区 -->
      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<style scoped>
.app-layout {
  min-height: 100vh;
}

.top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #e5e6eb;
  position: sticky;
  top: 0;
  z-index: 10;
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

.body-layout {
  min-height: calc(100vh - 56px);
}

.sidebar {
  background: #fff;
  border-right: 1px solid #e5e6eb;
  min-height: calc(100vh - 56px);
}

.content {
  padding: 24px;
  background: #f5f7fa;
}
</style>
