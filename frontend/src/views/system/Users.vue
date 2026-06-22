<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import {
  getUserListApi, createUserApi, updateUserApi, deleteUserApi,
  getRoleListApi,
  type UserItem, type RoleItem,
} from '@/api/system'
import { Message } from '@arco-design/web-vue'

// ——— 状态 ———
const loading = ref(false)
const users = ref<UserItem[]>([])
const roles = ref<RoleItem[]>([])
const total = ref(0)
const query = reactive({ keyword: '', page: 1, size: 10 })

const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref(0)
const form = reactive({ username: '', password: '', nickname: '', role_id: 0 })

// ——— 加载 ———
async function fetchUsers() {
  loading.value = true
  try {
    const res = await getUserListApi(query)
    if (res.data.data) {
      users.value = res.data.data.list
      total.value = res.data.data.total
    }
  } finally {
    loading.value = false
  }
}

async function fetchRoles() {
  const res = await getRoleListApi()
  if (res.data.data) roles.value = res.data.data
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})

// ——— 搜索/分页 ———
function handleSearch() {
  query.page = 1
  fetchUsers()
}
function handleReset() {
  query.keyword = ''
  query.page = 1
  fetchUsers()
}
function handlePageChange(page: number) {
  query.page = page
  fetchUsers()
}

// ——— 新增/编辑弹窗 ———
function openCreate() {
  modalMode.value = 'create'
  editingId.value = 0
  form.username = ''
  form.password = ''
  form.nickname = ''
  form.role_id = 0
  modalVisible.value = true
}

function openEdit(row: UserItem) {
  modalMode.value = 'edit'
  editingId.value = row.id
  form.username = row.username
  form.password = ''
  form.nickname = row.nickname
  form.role_id = row.role_id
  modalVisible.value = true
}

async function handleSubmit() {
  try {
    if (modalMode.value === 'create') {
      await createUserApi({ ...form })
      Message.success('创建成功')
    } else {
      await updateUserApi(editingId.value, { nickname: form.nickname, role_id: form.role_id })
      Message.success('更新成功')
    }
    modalVisible.value = false
    fetchUsers()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}

// ——— 删除 ———
async function handleDelete(row: UserItem) {
  try {
    await deleteUserApi(row.id)
    Message.success('删除成功')
    fetchUsers()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '删除失败')
  }
}

// ——— 列表 ———
const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '用户名', dataIndex: 'username', width: 140 },
  { title: '昵称', dataIndex: 'nickname', width: 140 },
  { title: '角色', slotName: 'role', width: 120 },
  { title: '创建时间', dataIndex: 'created_at', width: 160 },
  { title: '操作', slotName: 'actions', width: 180 },
]
</script>

<template>
  <div class="users-page">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <a-input v-model="query.keyword" placeholder="用户名/昵称" allow-clear style="width: 240px" />
      <a-button type="primary" @click="handleSearch">查询</a-button>
      <a-button @click="handleReset">重置</a-button>
      <a-button type="primary" @click="openCreate" style="margin-left: auto">+ 新增用户</a-button>
    </div>

    <!-- 表格 -->
    <a-table
      :columns="columns"
      :data="users"
      :loading="loading"
      :pagination="{ total, current: query.page, pageSize: query.size, showTotal: true }"
      @page-change="handlePageChange"
      row-key="id"
    >
      <template #role="{ record }">
        {{ record.role_name || (record.role_id === 0 ? '未分配' : `角色#${record.role_id}`) }}
      </template>
      <template #actions="{ record }">
        <a-space>
          <a-button type="text" size="small" @click="openEdit(record)">编辑</a-button>
          <a-popconfirm content="确定删除该用户？" @ok="handleDelete(record)">
            <a-button type="text" size="small" status="danger">删除</a-button>
          </a-popconfirm>
        </a-space>
      </template>
    </a-table>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="modalVisible" :title="modalMode === 'create' ? '新增用户' : '编辑用户'" @ok="handleSubmit">
      <a-form :model="form" layout="vertical">
        <a-form-item label="用户名" v-if="modalMode === 'create'">
          <a-input v-model="form.username" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="密码" v-if="modalMode === 'create'">
          <a-input-password v-model="form.password" placeholder="请输入密码" />
        </a-form-item>
        <a-form-item label="昵称">
          <a-input v-model="form.nickname" placeholder="请输入昵称" />
        </a-form-item>
        <a-form-item label="角色">
          <a-select v-model="form.role_id" placeholder="请选择角色" allow-clear>
            <a-option v-for="r in roles" :key="r.id" :value="r.id">{{ r.name }}</a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.users-page {
  max-width: 1200px;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}
</style>
