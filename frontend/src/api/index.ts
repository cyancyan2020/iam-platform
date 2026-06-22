import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
}

export interface ApiResponse<T = any> {
  code: number
  message: string
  data?: T
}

export const loginApi = (params: LoginParams) =>
  api.post<ApiResponse<LoginResult>>('/users/login', params)

export const registerApi = (params: {
  username: string
  password: string
  nickname?: string
}) => api.post<ApiResponse>('/users/register', params)

export const getProfileApi = () =>
  api.get<ApiResponse>('/profile')

export default api
