<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import {
  getPermissionListApi, createPermissionApi, updatePermissionApi, deletePermissionApi,
  type PermissionItem,
} from '@/api/system'
import { Message } from '@arco-design/web-vue'

// ——— 状态 ———
const loading = ref(false)
const permissions = ref<PermissionItem[]>([])

// 新增/编辑弹窗
const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref(0)
const form = reactive({ path: '', method: 'GET', name: '', parent_id: 0 })

// ——— 加载 ———
async function fetchPermissions() {
  loading.value = true
  try {
    const res = await getPermissionListApi()
    if (res.data.data) permissions.value = res.data.data
  } finally {
    loading.value = false
  }
}

onMounted(fetchPermissions)

// ——— 权限树结构 ———
function buildTree() {
  const menus = permissions.value.filter(p => p.method === 'MENU')
  return menus.map(menu => ({
    ...menu,
    children: permissions.value.filter(p => p.parent_id === menu.id && p.method !== 'MENU'),
  }))
}

// ——— 弹窗 ———
function openCreate(parentId = 0) {
  modalMode.value = 'create'
  editingId.value = 0
  form.path = ''
  form.method = 'GET'
  form.name = ''
  form.parent_id = parentId
  modalVisible.value = true
}

function openEdit(row: PermissionItem) {
  modalMode.value = 'edit'
  editingId.value = row.id
  form.path = row.path
  form.method = row.method
  form.name = row.name
  form.parent_id = row.parent_id
  modalVisible.value = true
}

async function handleSubmit() {
  try {
    if (modalMode.value === 'create') {
      await createPermissionApi({ ...form })
      Message.success('创建成功')
    } else {
      await updatePermissionApi(editingId.value, { ...form })
      Message.success('更新成功')
    }
    modalVisible.value = false
    fetchPermissions()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}

async function handleDelete(row: PermissionItem) {
  try {
    await deletePermissionApi(row.id)
    Message.success('删除成功')
    fetchPermissions()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '删除失败')
  }
}

// ——— 方法名颜色 ———
function methodColor(m: string) {
  const map: Record<string, string> = {
    GET: '#00b42a', POST: '#165dff', PUT: '#e6a23c', DELETE: '#f53f3f', MENU: '#86909c',
  }
  return map[m] || '#333'
}
</script>

<template>
  <div class="perms-page">
    <div class="search-bar">
      <a-button type="primary" @click="openCreate(0)">+ 新增菜单</a-button>
    </div>

    <div class="perm-tree">
      <template v-for="menu in buildTree()" :key="menu.id">
        <!-- 菜单节点 -->
        <div class="perm-group">
          <div class="perm-group-header">
            <span class="group-icon">&#128193;</span>
            <span class="group-name">{{ menu.name }}</span>
            <a-space size="mini">
              <a-button type="text" size="mini" @click="openCreate(menu.id)">+ 子权限</a-button>
              <a-button type="text" size="mini" @click="openEdit(menu)">编辑</a-button>
              <a-popconfirm content="删除菜单将同时无法访问其下权限，确定删除？" @ok="handleDelete(menu)">
                <a-button type="text" size="mini" status="danger">删除</a-button>
              </a-popconfirm>
            </a-space>
          </div>

          <!-- 子权限 -->
          <div class="perm-children">
            <div v-for="child in menu.children" :key="child.id" class="perm-item">
              <span class="perm-method" :style="{ color: methodColor(child.method) }">{{ child.method }}</span>
              <span class="perm-path">{{ child.path }}</span>
              <span class="perm-name">{{ child.name }}</span>
              <a-space size="mini" style="margin-left: auto">
                <a-button type="text" size="mini" @click="openEdit(child)">编辑</a-button>
                <a-popconfirm content="确定删除？" @ok="handleDelete(child)">
                  <a-button type="text" size="mini" status="danger">删除</a-button>
                </a-popconfirm>
              </a-space>
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="modalVisible" :title="modalMode === 'create' ? '新增权限' : '编辑权限'" @ok="handleSubmit">
      <a-form :model="form" layout="vertical">
        <a-form-item label="权限名称">
          <a-input v-model="form.name" placeholder="如 用户列表" />
        </a-form-item>
        <a-form-item label="请求路径">
          <a-input v-model="form.path" placeholder="如 /api/v1/users" />
        </a-form-item>
        <a-form-item label="请求方法">
          <a-select v-model="form.method">
            <a-option value="MENU">MENU（菜单）</a-option>
            <a-option value="GET">GET</a-option>
            <a-option value="POST">POST</a-option>
            <a-option value="PUT">PUT</a-option>
            <a-option value="DELETE">DELETE</a-option>
          </a-select>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style scoped>
.perms-page {
  max-width: 960px;
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

.perm-tree {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.perm-group {
  margin-bottom: 12px;
  border: 1px solid #e5e6eb;
  border-radius: 6px;
}

.perm-group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #f7f8fa;
  border-bottom: 1px solid #e5e6eb;
}

.group-icon {
  font-size: 16px;
}

.group-name {
  font-weight: 600;
  font-size: 14px;
}

.perm-children {
  padding: 4px 0;
}

.perm-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px 8px 48px;
  font-size: 13px;
}

.perm-method {
  font-family: monospace;
  font-weight: 600;
  width: 52px;
  font-size: 12px;
}

.perm-path {
  font-family: monospace;
  color: #4e5969;
  width: 280px;
}

.perm-name {
  color: #1d2129;
}
</style>
