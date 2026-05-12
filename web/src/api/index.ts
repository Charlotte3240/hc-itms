import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api',
  timeout: 300000,
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
    } else {
      ElMessage.error(error.response?.data?.error || '请求失败')
    }
    return Promise.reject(error)
  }
)

export const authApi = {
  login: (username: string, password: string) =>
    api.post('/auth/login', { username, password }),
  register: (username: string, password: string) =>
    api.post('/auth/register', { username, password }),
}

export const appApi = {
  list: (platform?: string) =>
    api.get('/apps', { params: { platform } }),
  get: (id: number) => api.get(`/apps/${id}`),
  update: (id: number, data: any) => api.patch(`/apps/${id}`, data),
  delete: (id: number) => api.delete(`/apps/${id}`),
  uploadVersion: (id: number, file: File, changelog: string, onProgress?: (e: any) => void) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('changelog', changelog)
    return api.post(`/apps/${id}/versions`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      onUploadProgress: onProgress,
    })
  },
  deleteVersion: (id: number) => api.delete(`/versions/${id}`),
}

export const downloadApi = {
  getLatest: (appId: number) => axios.get(`/d/${appId}/latest`),
}

export default api
