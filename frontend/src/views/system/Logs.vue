<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import api, { type ApiResponse } from '@/api'
import { Message } from '@arco-design/web-vue'

// ——— 类型 ———
interface LogItem {
  id: number
  user_id: number
  username: string
  method: string
  path: string
  ip: string
  user_agent: string
  status_code: number
  duration_ms: number
  created_at: string
}

interface LogListResult {
  list: LogItem[]
  total: number
}

// ——— 状态 ———
const loading = ref(false)
const logs = ref<LogItem[]>([])
const total = ref(0)
const query = reactive({
  start: '',
  end: '',
  page: 1,
  size: 20,
})
const dateRange = ref<string[]>([])

function fmtDate(d: Date) {
  const pad = (n: number) => n.toString().padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

// ——— 加载 ———
async function fetchLogs() {
  loading.value = true
  try {
    const res = await api.get<ApiResponse<LogListResult>>('/logs', { params: query })
    if (res.data.data) {
      logs.value = res.data.data.list
      total.value = res.data.data.total
    }
  } catch {
    Message.error('加载日志失败')
  } finally {
    loading.value = false
  }
}

onMounted(fetchLogs)

// ——— 搜索/分页 ———
function handleSearch() {
  if (dateRange.value.length === 2) {
    query.start = fmtDate(new Date(dateRange.value[0]))
    query.end = fmtDate(new Date(dateRange.value[1]))
  }
  query.page = 1
  fetchLogs()
}
function handleReset() {
  query.start = ''
  query.end = ''
  dateRange.value = []
  query.page = 1
  fetchLogs()
}
function handlePageChange(page: number) {
  query.page = page
  fetchLogs()
}

// ——— CSV 导出 ———
function handleExport() {
  const header = ['ID', '用户ID', '用户名', '方法', '路径', 'IP', '状态码', '耗时ms', '时间']
  const rows = logs.value.map(r => [
    r.id, r.user_id, r.username, r.method, r.path, r.ip, r.status_code, r.duration_ms, r.created_at,
  ])
  const csv = [header, ...rows].map(row => row.map(v => `"${v ?? ''}"`).join(',')).join('\n')
  const blob = new Blob(['﻿' + csv], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `操作日志_${new Date().toISOString().slice(0, 10)}.csv`
  a.click()
  URL.revokeObjectURL(url)
}

function methodColor(m: string) {
  const map: Record<string, string> = {
    GET: '#00b42a', POST: '#165dff', PUT: '#e6a23c', DELETE: '#f53f3f',
  }
  return map[m] || '#333'
}

function statusColor(code: number) {
  if (code < 300) return '#00b42a'
  if (code < 400) return '#e6a23c'
  return '#f53f3f'
}

// ——— 列表 ———
const columns = [
  { title: 'ID', dataIndex: 'id', width: 70 },
  { title: '用户', slotName: 'user', width: 110 },
  { title: '方法', slotName: 'method', width: 70 },
  { title: '路径', dataIndex: 'path', width: 240, ellipsis: true },
  { title: 'IP', dataIndex: 'ip', width: 130 },
  { title: '状态', slotName: 'status', width: 70 },
  { title: '耗时', slotName: 'duration', width: 80 },
  { title: '时间', dataIndex: 'created_at', width: 160 },
]
</script>

<template>
  <div class="logs-page">
    <!-- 筛选栏 -->
    <div class="search-bar">
      <a-range-picker v-model="dateRange" show-time format="YYYY-MM-DD HH:mm:ss" style="width: 380px" />
      <a-button type="primary" @click="handleSearch">查询</a-button>
      <a-button @click="handleReset">重置</a-button>
      <a-button @click="handleExport" style="margin-left: auto">导出 CSV</a-button>
    </div>

    <!-- 表格 -->
    <a-table
      :columns="columns"
      :data="logs"
      :loading="loading"
      :pagination="{ total, current: query.page, pageSize: query.size, showTotal: true }"
      @page-change="handlePageChange"
      row-key="id"
      size="small"
    >
      <template #user="{ record }">
        {{ record.username || `#${record.user_id}` }}
      </template>
      <template #method="{ record }">
        <span :style="{ color: methodColor(record.method), fontWeight: 600, fontFamily: 'monospace' }">
          {{ record.method }}
        </span>
      </template>
      <template #status="{ record }">
        <span :style="{ color: statusColor(record.status_code), fontWeight: 600 }">
          {{ record.status_code }}
        </span>
      </template>
      <template #duration="{ record }">
        {{ record.duration_ms }}ms
      </template>
    </a-table>
  </div>
</template>

<style scoped>
.logs-page {
  max-width: 1400px;
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
