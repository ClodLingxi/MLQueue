import axios from 'axios'

// V2 API 基础配置
const API_BASE_URL = import.meta.env.VITE_API_V2_URL || 'http://localhost:8080/v2'
const DEFAULT_API_KEY = import.meta.env.VITE_API_KEY || ''

console.log('MLQueue V2 API 配置:')
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
    console.error('V2 API Error:', errorMessage)
    return Promise.reject(error)
  }
)

// ========== 组（Group）管理 API ==========

/**
 * 创建组
 * @param {string} name - 组名称
 * @param {string} description - 组描述
 */
export const createGroup = (name, description = '') => {
  return apiClient.post('/groups', { name, description })
}

/**
 * 列出所有组
 */
export const listGroups = () => {
  return apiClient.get('/groups')
}

/**
 * 获取组详情
 * @param {string} groupId - 组ID
 */
export const getGroup = (groupId) => {
  return apiClient.get(`/groups/${groupId}`)
}

/**
 * 更新组
 * @param {string} groupId - 组ID
 * @param {object} data - 更新数据 {name, description}
 */
export const updateGroup = (groupId, data) => {
  return apiClient.put(`/groups/${groupId}`, data)
}

/**
 * 删除组
 * @param {string} groupId - 组ID
 */
export const deleteGroup = (groupId) => {
  return apiClient.delete(`/groups/${groupId}`)
}

// ========== 训练单元（Training Unit）管理 API ==========

/**
 * 创建训练单元
 * @param {string} groupId - 组ID
 * @param {object} data - {name, description, config}
 */
export const createTrainingUnit = (groupId, data) => {
  return apiClient.post(`/groups/${groupId}/units`, data)
}

/**
 * 列出训练单元
 * @param {string} groupId - 组ID
 */
export const listTrainingUnits = (groupId) => {
  return apiClient.get(`/groups/${groupId}/units`)
}

/**
 * 获取训练单元详情
 * @param {string} unitId - 训练单元ID
 */
export const getTrainingUnit = (unitId) => {
  return apiClient.get(`/units/${unitId}`)
}

/**
 * 同步训练单元（Python 客户端使用）
 * @param {string} unitId - 训练单元ID
 * @param {number} clientVersion - 客户端版本号
 */
export const syncTrainingUnit = (unitId, clientVersion = 0) => {
  return apiClient.post(`/units/${unitId}/sync`, { client_version: clientVersion })
}

/**
 * 更新训练单元
 * @param {string} unitId - 训练单元ID
 * @param {object} data - {name, description, config}
 */
export const updateTrainingUnit = (unitId, data) => {
  return apiClient.put(`/units/${unitId}`, data)
}

/**
 * 删除训练单元
 * @param {string} unitId - 训练单元ID
 */
export const deleteTrainingUnit = (unitId) => {
  return apiClient.delete(`/units/${unitId}`)
}

/**
 * 发送心跳保持连接（Python 客户端使用）
 * @param {string} unitId - 训练单元ID
 * @note Python客户端应每隔5-8秒调用一次此接口
 */
export const sendHeartbeat = (unitId) => {
  return apiClient.post(`/units/${unitId}/heartbeat`)
}

// ========== 训练队列（Training Queue）管理 API ==========

/**
 * 创建训练队列
 * @param {string} unitId - 训练单元ID
 * @param {object} data - {name, parameters, created_by}
 * @note order字段由后端自动分配，新队列追加到末尾
 */
export const createTrainingQueue = (unitId, data) => {
  return apiClient.post(`/units/${unitId}/queues`, {
    ...data,
    created_by: data.created_by || 'web'
  })
}

/**
 * 批量创建训练队列
 * @param {string} unitId - 训练单元ID
 * @param {array} queues - 队列数组
 * @param {string} createdBy - 创建者 ('web' 或 'client')
 */
export const batchCreateTrainingQueues = (unitId, queues, createdBy = 'web') => {
  return apiClient.post(`/units/${unitId}/queues/batch`, {
    queues,
    created_by: createdBy
  })
}

/**
 * 列出训练队列
 * @param {string} unitId - 训练单元ID
 * @param {string} status - 可选，过滤状态 (pending/running/completed/failed/cancelled)
 */
export const listTrainingQueues = (unitId, status = '') => {
  const params = status ? { status } : {}
  return apiClient.get(`/units/${unitId}/queues`, { params })
}

/**
 * 获取队列详情
 * @param {string} queueId - 队列ID
 */
export const getTrainingQueue = (queueId) => {
  return apiClient.get(`/queues/${queueId}`)
}

/**
 * 更新训练队列（仅 pending 状态）
 * @param {string} queueId - 队列ID
 * @param {object} data - {name, parameters}
 * @note order字段不能通过此接口修改，请使用 reorderTrainingQueues
 */
export const updateTrainingQueue = (queueId, data) => {
  return apiClient.put(`/queues/${queueId}`, data)
}

/**
 * 删除训练队列（不能删除 running 状态）
 * @param {string} queueId - 队列ID
 */
export const deleteTrainingQueue = (queueId) => {
  return apiClient.delete(`/queues/${queueId}`)
}

/**
 * 重新排序训练队列
 * @param {string} unitId - 训练单元ID
 * @param {array} queueIds - 队列ID数组（按新的执行顺序排列）
 * @note 只能调整pending状态的队列，系统会自动重新分配order值
 */
export const reorderTrainingQueues = (unitId, queueIds) => {
  return apiClient.post(`/units/${unitId}/queues/reorder`, { queue_ids: queueIds })
}

// ========== Python 客户端专用 API ==========

/**
 * 开始执行队列（Python 客户端调用）
 * @param {string} queueId - 队列ID
 */
export const startTrainingQueue = (queueId) => {
  return apiClient.post(`/queues/${queueId}/start`)
}

/**
 * 标记队列完成（Python 客户端调用）
 * @param {string} queueId - 队列ID
 * @param {object} result - 训练结果
 * @param {object} metrics - 训练指标
 */
export const completeTrainingQueue = (queueId, result, metrics = {}) => {
  return apiClient.post(`/queues/${queueId}/complete`, { result, metrics })
}

/**
 * 标记队列失败（Python 客户端调用）
 * @param {string} queueId - 队列ID
 * @param {string} errorMsg - 错误消息
 */
export const failTrainingQueue = (queueId, errorMsg) => {
  return apiClient.post(`/queues/${queueId}/fail`, { error_msg: errorMsg })
}

// ========== 统计信息 API ==========

/**
 * 获取组的统计信息
 * @param {string} groupId - 组ID
 */
export const getGroupStatistics = (groupId) => {
  return apiClient.get(`/groups/${groupId}/statistics`)
}

/**
 * 获取训练单元的统计信息
 * @param {string} unitId - 训练单元ID
 */
export const getUnitStatistics = (unitId) => {
  return apiClient.get(`/units/${unitId}/statistics`)
}

// ========== 工具函数 ==========

/**
 * 格式化队列状态为中文
 */
export const formatQueueStatus = (status) => {
  const statusMap = {
    pending: '待执行',
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return statusMap[status] || status
}

/**
 * 获取状态对应的颜色
 */
export const getStatusColor = (status) => {
  const colorMap = {
    pending: '#e6a23c',
    running: '#409eff',
    completed: '#67c23a',
    failed: '#f56c6c',
    cancelled: '#909399',
    idle: '#909399'
  }
  return colorMap[status] || '#909399'
}

// 导出所有 API
export default {
  // 组管理
  createGroup,
  listGroups,
  getGroup,
  updateGroup,
  deleteGroup,

  // 训练单元管理
  createTrainingUnit,
  listTrainingUnits,
  getTrainingUnit,
  syncTrainingUnit,
  updateTrainingUnit,
  deleteTrainingUnit,
  sendHeartbeat,

  // 训练队列管理
  createTrainingQueue,
  batchCreateTrainingQueues,
  listTrainingQueues,
  getTrainingQueue,
  updateTrainingQueue,
  deleteTrainingQueue,
  reorderTrainingQueues,

  // Python 客户端专用
  startTrainingQueue,
  completeTrainingQueue,
  failTrainingQueue,

  // 统计信息
  getGroupStatistics,
  getUnitStatistics,

  // 工具函数
  formatQueueStatus,
  getStatusColor
}
