import api, { type ApiResponse } from './index'

// ---------- 用户 ----------

export interface UserItem {
  id: number
  username: string
  nickname: string
  role_id: number
  role_name: string
  created_at: string
}

export interface UserListResult {
  list: UserItem[]
  total: number
}

export interface UserListParams {
  keyword?: string
  page?: number
  size?: number
}

export const getUserListApi = (params: UserListParams) =>
  api.get<ApiResponse<UserListResult>>('/users', { params })

export const createUserApi = (data: {
  username: string
  password: string
  nickname?: string
  role_id?: number
}) => api.post<ApiResponse>('/users', data)

export const updateUserApi = (id: number, data: { nickname?: string; role_id?: number }) =>
  api.put<ApiResponse>(`/users/${id}`, data)

export const deleteUserApi = (id: number) =>
  api.delete<ApiResponse>(`/users/${id}`)

export const assignRoleApi = (userId: number, roleId: number) =>
  api.post<ApiResponse>(`/users/${userId}/role`, { role_id: roleId })

// ---------- 角色 ----------

export interface RoleItem {
  id: number
  code: string
  name: string
  created_at: string
}

export const getRoleListApi = () =>
  api.get<ApiResponse<RoleItem[]>>('/roles')

export const createRoleApi = (data: { code: string; name: string }) =>
  api.post<ApiResponse<RoleItem>>('/roles', data)

export const updateRoleApi = (id: number, data: { code: string; name: string }) =>
  api.put<ApiResponse>(`/roles/${id}`, data)

export const deleteRoleApi = (id: number) =>
  api.delete<ApiResponse>(`/roles/${id}`)

export const setRolePermissionsApi = (roleId: number, permissionIds: number[]) =>
  api.post<ApiResponse>(`/roles/${roleId}/permissions`, { permission_ids: permissionIds })

// ---------- 权限 ----------

export interface PermissionItem {
  id: number
  path: string
  method: string
  name: string
  parent_id: number
  created_at: string
}

export const getPermissionListApi = () =>
  api.get<ApiResponse<PermissionItem[]>>('/permissions')

export const createPermissionApi = (data: {
  path: string
  method: string
  name: string
  parent_id?: number
}) => api.post<ApiResponse<PermissionItem>>('/permissions', data)

export const updatePermissionApi = (id: number, data: {
  path: string
  method: string
  name: string
  parent_id?: number
}) => api.put<ApiResponse>(`/permissions/${id}`, data)

export const deletePermissionApi = (id: number) =>
  api.delete<ApiResponse>(`/permissions/${id}`)
