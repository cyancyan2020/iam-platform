<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import {
  getRoleListApi, createRoleApi, updateRoleApi, deleteRoleApi,
  getPermissionListApi, setRolePermissionsApi,
  type RoleItem, type PermissionItem,
} from '@/api/system'
import { Message } from '@arco-design/web-vue'
import { IconEdit } from '@arco-design/web-vue/es/icon'

// ——— 状态 ———
const loading = ref(false)
const roles = ref<RoleItem[]>([])
const permissions = ref<PermissionItem[]>([])

// 新增/编辑弹窗
const modalVisible = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref(0)
const form = reactive({ code: '', name: '' })

// 权限分配抽屉
const drawerVisible = ref(false)
const drawerRoleId = ref(0)
const drawerRoleName = ref('')
const checkedPermIds = ref<number[]>([])

// ——— 加载 ———
async function fetchRoles() {
  loading.value = true
  try {
    const res = await getRoleListApi()
    if (res.data.data) roles.value = res.data.data
  } finally {
    loading.value = false
  }
}

async function fetchPermissions() {
  const res = await getPermissionListApi()
  if (res.data.data) permissions.value = res.data.data
}

onMounted(() => {
  fetchRoles()
  fetchPermissions()
})

// ——— 弹窗 ———
function openCreate() {
  modalMode.value = 'create'
  editingId.value = 0
  form.code = ''
  form.name = ''
  modalVisible.value = true
}

function openEdit(row: RoleItem) {
  modalMode.value = 'edit'
  editingId.value = row.id
  form.code = row.code
  form.name = row.name
  modalVisible.value = true
}

async function handleSubmit() {
  try {
    if (modalMode.value === 'create') {
      await createRoleApi({ ...form })
      Message.success('创建成功')
    } else {
      await updateRoleApi(editingId.value, { ...form })
      Message.success('更新成功')
    }
    modalVisible.value = false
    fetchRoles()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '操作失败')
  }
}

async function handleDelete(row: RoleItem) {
  try {
    await deleteRoleApi(row.id)
    Message.success('删除成功')
    fetchRoles()
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '删除失败')
  }
}

// ——— 权限抽屉 ———
function openPermissionDrawer(row: RoleItem) {
  drawerRoleId.value = row.id
  drawerRoleName.value = row.name
  // TODO: 从后端获取角色当前的权限列表
  checkedPermIds.value = []
  drawerVisible.value = true
}

async function handleSavePermissions() {
  try {
    await setRolePermissionsApi(drawerRoleId.value, checkedPermIds.value)
    Message.success('权限分配成功')
    drawerVisible.value = false
  } catch (e: any) {
    Message.error(e?.response?.data?.message || '保存失败')
  }
}

// 构建权限树（按 parent_id 分组）
function buildPermTree() {
  const root = permissions.value.filter(p => p.parent_id === 0 && p.method !== 'MENU')
  return root.map(r => ({
    ...r,
    children: permissions.value.filter(p => p.parent_id === r.id),
  }))
}

const columns = [
  { title: 'ID', dataIndex: 'id', width: 80 },
  { title: '编码', dataIndex: 'code', width: 120 },
  { title: '名称', dataIndex: 'name', width: 160 },
  { title: '操作', slotName: 'actions', width: 260 },
]
</script>

<template>
  <div class="roles-page">
    <div class="search-bar">
      <a-button type="primary" @click="openCreate">+ 新增角色</a-button>
    </div>

    <a-table :columns="columns" :data="roles" :loading="loading" row-key="id">
      <template #actions="{ record }">
        <a-space>
          <a-button type="text" size="small" @click="openEdit(record)">编辑</a-button>
          <a-button type="text" size="small" @click="openPermissionDrawer(record)">
            <template #icon><icon-edit /></template>
            权限
          </a-button>
          <a-popconfirm content="确定删除该角色？" @ok="handleDelete(record)">
            <a-button type="text" size="small" status="danger">删除</a-button>
          </a-popconfirm>
        </a-space>
      </template>
    </a-table>

    <!-- 新增/编辑弹窗 -->
    <a-modal v-model:visible="modalVisible" :title="modalMode === 'create' ? '新增角色' : '编辑角色'" @ok="handleSubmit">
      <a-form :model="form" layout="vertical">
        <a-form-item label="角色编码">
          <a-input v-model="form.code" placeholder="如 editor" />
        </a-form-item>
        <a-form-item label="角色名称">
          <a-input v-model="form.name" placeholder="如 编辑者" />
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 权限分配抽屉 -->
    <a-drawer v-model:visible="drawerVisible" :title="`分配权限 - ${drawerRoleName}`" :width="480" :footer="false">
      <div>
        <a-checkbox-group v-model="checkedPermIds" direction="vertical">
          <template v-for="p in buildPermTree()" :key="p.id">
            <div style="font-weight: 600; margin: 12px 0 4px">{{ p.name }}</div>
            <a-checkbox v-for="child in p.children" :key="child.id" :value="child.id">
              {{ child.name }} <span style="color: #999; font-size: 12px">{{ child.method }} {{ child.path }}</span>
            </a-checkbox>
          </template>
        </a-checkbox-group>
      </div>
      <div style="margin-top: 24px">
        <a-button type="primary" long @click="handleSavePermissions">保存权限</a-button>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped>
.roles-page {
  max-width: 900px;
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
