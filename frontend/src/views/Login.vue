<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { Message } from '@arco-design/web-vue'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

const rules = {
  username: [
    { required: true, message: '请输入用户名' },
    { minLength: 3, message: '用户名至少 3 个字符' },
  ],
  password: [
    { required: true, message: '请输入密码' },
    { minLength: 6, message: '密码至少 6 个字符' },
  ],
}

async function handleSubmit() {
  loading.value = true
  try {
    await userStore.login({ username: form.username, password: form.password })
    Message.success('登录成功')
    router.replace('/dashboard')
  } catch (e: any) {
    const msg = e?.response?.data?.message || '登录失败，请重试'
    Message.error(msg)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-wrapper">
    <div class="login-card">
      <div class="login-header">
        <div class="logo-mark">I</div>
        <h1 class="platform-title">IAM Platform</h1>
        <p class="platform-desc">统一身份认证与访问管理</p>
      </div>

      <a-form
        :model="form"
        :rules="rules"
        layout="vertical"
        size="large"
        @submit="handleSubmit"
        class="login-form"
      >
        <a-form-item field="username">
          <a-input
            v-model="form.username"
            placeholder="用户名"
            allow-clear
          >
            <template #prefix>
              <icon-user />
            </template>
          </a-input>
        </a-form-item>

        <a-form-item field="password">
          <a-input-password
            v-model="form.password"
            placeholder="密码"
            allow-clear
            @keyup.enter="handleSubmit"
          >
            <template #prefix>
              <icon-lock />
            </template>
          </a-input-password>
        </a-form-item>

        <a-button
          type="primary"
          html-type="submit"
          long
          :loading="loading"
          class="login-btn"
        >
          登 录
        </a-button>
      </a-form>

      <div class="login-footer">
        <span>首次使用？请联系管理员分配账号</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f0f5ff 0%, #e8f0fe 50%, #f5f7fa 100%);
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 48px 40px 36px;
  box-shadow:
    0 2px 8px rgba(0, 0, 0, 0.04),
    0 8px 24px rgba(0, 0, 0, 0.06);
}

.login-header {
  text-align: center;
  margin-bottom: 36px;
}

.logo-mark {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #165dff, #4080ff);
  color: #fff;
  font-size: 22px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16px;
  letter-spacing: 0;
}

.platform-title {
  font-size: 22px;
  font-weight: 600;
  color: #1d2129;
  margin-bottom: 6px;
  letter-spacing: -0.3px;
}

.platform-desc {
  font-size: 14px;
  color: #86909c;
  margin: 0;
}

.login-form {
  margin-top: 8px;
}

.login-form :deep(.arco-form-item) {
  margin-bottom: 20px;
}

.login-form :deep(.arco-input-wrapper) {
  border-radius: 8px;
}

.login-btn {
  margin-top: 8px;
  border-radius: 8px;
  height: 44px;
  font-size: 15px;
  letter-spacing: 4px;
}

.login-footer {
  text-align: center;
  margin-top: 28px;
  font-size: 13px;
  color: #c9cdd4;
}
</style>
