import axios from 'axios'

// API 基础配置 - 从环境变量读取，如果没有则使用默认值
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/v1'

// 默认 API Key（可选，优先使用 localStorage 中的 token）
const DEFAULT_API_KEY = import.meta.env.VITE_API_KEY || ''

console.log('MLQueue API 配置:')
console.log('- API Base URL:', API_BASE_URL)
console.log('- 使用默认 API Key:', DEFAULT_API_KEY ? '是' : '否')

// 创建 axios 实例
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器 - 添加 Bearer Token
apiClient.interceptors.request.use(
  config => {
    // 优先使用 localStorage 中的 token，其次使用环境变量中的默认 key
    const token = localStorage.getItem('mlqueue_token') || DEFAULT_API_KEY
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器 - 统一处理响应
apiClient.interceptors.response.use(
  response => response.data,
  error => {
    const errorMessage = error.response?.data?.error || error.message
    console.error('API Error:', errorMessage)
    return Promise.reject(error)
  }
)

// ========== 任务管理 API ==========

// 创建训练任务
export const createTask = (taskData) => {
  return apiClient.post('/tasks', taskData)
}

// 批量创建任务
export const batchCreateTasks = (tasks) => {
  return apiClient.post('/tasks/batch', { tasks })
}

// 获取任务信息
export const getTask = (taskId) => {
  return apiClient.get(`/tasks/${taskId}`)
}

// 列出任务
export const listTasks = (params = {}) => {
  return apiClient.get('/tasks', { params })
}

// 更新任务优先级
export const updateTaskPriority = (taskId, priority) => {
  return apiClient.patch(`/tasks/${taskId}/priority`, { priority })
}

// 取消任务
export const cancelTask = (taskId, reason = '') => {
  return apiClient.post(`/tasks/${taskId}/cancel`, { reason })
}

// 上传训练结果
export const uploadTaskResult = (taskId, result, artifacts = {}) => {
  return apiClient.post(`/tasks/${taskId}/result`, { result, artifacts })
}

// 获取任务日志
export const getTaskLogs = (taskId, lines = 100) => {
  return apiClient.get(`/tasks/${taskId}/logs`, { params: { lines } })
}

// ========== 队列管理 API ==========

// 获取队列状态
export const getQueueStatus = () => {
  return apiClient.get('/queue/status')
}

// 重新排列队列
export const reorderQueue = (taskIds) => {
  return apiClient.post('/queue/reorder', { task_ids: taskIds })
}

// 暂停队列
export const pauseQueue = () => {
  return apiClient.post('/queue/pause')
}

// 恢复队列
export const resumeQueue = () => {
  return apiClient.post('/queue/resume')
}

// ========== 配置管理 API ==========

// 获取配置模板
export const getConfigTemplates = () => {
  return apiClient.get('/configs/templates')
}

// 创建配置模板
export const createConfigTemplate = (name, config, description = '') => {
  return apiClient.post('/configs/templates', { name, config, description })
}

// ========== 统计和监控 API ==========

// 获取任务统计
export const getTaskStatistics = (startDate, endDate) => {
  const params = {}
  if (startDate) params.start_date = startDate
  if (endDate) params.end_date = endDate
  return apiClient.get('/statistics/tasks', { params })
}

// ========== Token 管理 ==========

// 设置 API Token
export const setApiToken = (token) => {
  localStorage.setItem('mlqueue_token', token)
}

// 获取 API Token
export const getApiToken = () => {
  return localStorage.getItem('mlqueue_token')
}

// 清除 API Token
export const clearApiToken = () => {
  localStorage.removeItem('mlqueue_token')
}

export default {
  // 任务管理
  createTask,
  batchCreateTasks,
  getTask,
  listTasks,
  updateTaskPriority,
  cancelTask,
  uploadTaskResult,
  getTaskLogs,

  // 队列管理
  getQueueStatus,
  reorderQueue,
  pauseQueue,
  resumeQueue,

  // 配置管理
  getConfigTemplates,
  createConfigTemplate,

  // 统计和监控
  getTaskStatistics,

  // Token 管理
  setApiToken,
  getApiToken,
  clearApiToken
}
