<script setup>
import { ref, onMounted } from 'vue'
import { listTasks, cancelTask, updateTaskPriority } from '../api/mlqueue.js'

const emit = defineEmits(['view-detail'])

// 任务列表数据
const tasks = ref([])
const loading = ref(false)
const error = ref(null)

// 筛选和分页参数
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const totalTasks = ref(0)

// 状态选项
const statusOptions = [
  { value: '', label: '全部' },
  { value: 'pending', label: '待处理' },
  { value: 'queued', label: '队列中' },
  { value: 'running', label: '运行中' },
  { value: 'completed', label: '已完成' },
  { value: 'failed', label: '失败' },
  { value: 'cancelled', label: '已取消' }
]

// 状态颜色映射
const statusColors = {
  pending: '#909399',
  queued: '#e6a23c',
  running: '#409eff',
  completed: '#67c23a',
  failed: '#f56c6c',
  cancelled: '#909399'
}

// 获取任务列表
const fetchTasks = async () => {
  loading.value = true
  error.value = null
  try {
    const params = {
      limit: pageSize.value,
      offset: (currentPage.value - 1) * pageSize.value,
      sort: 'created_at'
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }

    const response = await listTasks(params)
    tasks.value = response.tasks || []
    totalTasks.value = response.total || 0
  } catch (err) {
    error.value = err.message || '获取任务列表失败'
    console.error('获取任务列表失败:', err)
  } finally {
    loading.value = false
  }
}

// 取消任务
const handleCancelTask = async (taskId) => {
  if (!confirm('确定要取消这个任务吗?')) return

  try {
    await cancelTask(taskId, '用户手动取消')
    await fetchTasks()
  } catch (err) {
    alert('取消任务失败: ' + err.message)
  }
}

// 更新优先级
const handleUpdatePriority = async (taskId, currentPriority) => {
  const newPriority = prompt('请输入新的优先级(数字越大优先级越高):', currentPriority)
  if (newPriority === null) return

  const priority = parseInt(newPriority)
  if (isNaN(priority)) {
    alert('优先级必须是数字')
    return
  }

  try {
    await updateTaskPriority(taskId, priority)
    await fetchTasks()
  } catch (err) {
    alert('更新优先级失败: ' + err.message)
  }
}

// 查看详情
const handleViewDetail = (taskId) => {
  emit('view-detail', taskId)
}

// 刷新列表
const handleRefresh = () => {
  fetchTasks()
}

// 状态改变时重置页码
const handleStatusChange = () => {
  currentPage.value = 1
  fetchTasks()
}

// 页码改变
const handlePageChange = (page) => {
  currentPage.value = page
  fetchTasks()
}

// 格式化时间
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 组件挂载时加载数据
onMounted(() => {
  fetchTasks()
})
</script>

<template>
  <div class="task-list">
    <div class="header">
      <h2>任务列表</h2>
      <div class="actions">
        <select v-model="statusFilter" @change="handleStatusChange" class="filter-select">
          <option v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
        <button @click="handleRefresh" class="btn-refresh">刷新</button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading">
      加载中...
    </div>

    <!-- 任务表格 -->
    <div v-else-if="tasks.length > 0" class="table-container">
      <table class="task-table">
        <thead>
          <tr>
            <th>任务ID</th>
            <th>名称</th>
            <th>状态</th>
            <th>优先级</th>
            <th>创建时间</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="task in tasks" :key="task.task_id">
            <td>
              <code class="task-id">{{ task.task_id }}</code>
            </td>
            <td>{{ task.name }}</td>
            <td>
              <span
                class="status-badge"
                :style="{ backgroundColor: statusColors[task.status] }"
              >
                {{ statusOptions.find(s => s.value === task.status)?.label || task.status }}
              </span>
            </td>
            <td>
              <span class="priority">{{ task.priority }}</span>
            </td>
            <td>{{ formatDate(task.created_at) }}</td>
            <td>
              <div class="action-buttons">
                <button @click="handleViewDetail(task.task_id)" class="btn-action">
                  详情
                </button>
                <button
                  v-if="task.status === 'pending' || task.status === 'queued'"
                  @click="handleUpdatePriority(task.task_id, task.priority)"
                  class="btn-action"
                >
                  优先级
                </button>
                <button
                  v-if="task.status === 'pending' || task.status === 'queued' || task.status === 'running'"
                  @click="handleCancelTask(task.task_id)"
                  class="btn-action btn-danger"
                >
                  取消
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 分页 -->
      <div class="pagination">
        <button
          @click="handlePageChange(currentPage - 1)"
          :disabled="currentPage === 1"
          class="page-btn"
        >
          上一页
        </button>
        <span class="page-info">
          第 {{ currentPage }} 页 / 共 {{ Math.ceil(totalTasks / pageSize) }} 页
          (总计 {{ totalTasks }} 个任务)
        </span>
        <button
          @click="handlePageChange(currentPage + 1)"
          :disabled="currentPage >= Math.ceil(totalTasks / pageSize)"
          class="page-btn"
        >
          下一页
        </button>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      暂无任务
    </div>
  </div>
</template>

<style scoped>
.task-list {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  margin: 0;
  font-size: 20px;
  color: #303133;
}

.actions {
  display: flex;
  gap: 10px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
}

.filter-select:focus {
  outline: none;
  border-color: #409eff;
}

.btn-refresh {
  padding: 8px 16px;
  background: #409eff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background 0.3s;
}

.btn-refresh:hover {
  background: #66b1ff;
}

.error-message {
  padding: 12px;
  background: #fef0f0;
  color: #f56c6c;
  border-radius: 4px;
  margin-bottom: 20px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #909399;
}

.table-container {
  overflow-x: auto;
}

.task-table {
  width: 100%;
  border-collapse: collapse;
}

.task-table th,
.task-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ebeef5;
}

.task-table th {
  background: #f5f7fa;
  color: #606266;
  font-weight: 600;
  font-size: 14px;
}

.task-table td {
  color: #606266;
  font-size: 14px;
}

.task-id {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-family: monospace;
}

.status-badge {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  color: white;
  font-size: 12px;
}

.priority {
  font-weight: 600;
  color: #409eff;
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.btn-action {
  padding: 4px 12px;
  border: 1px solid #dcdfe6;
  background: white;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.3s;
}

.btn-action:hover {
  border-color: #409eff;
  color: #409eff;
}

.btn-danger {
  color: #f56c6c;
  border-color: #f56c6c;
}

.btn-danger:hover {
  background: #f56c6c;
  color: white;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 20px;
  margin-top: 20px;
}

.page-btn {
  padding: 8px 16px;
  border: 1px solid #dcdfe6;
  background: white;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.page-btn:hover:not(:disabled) {
  border-color: #409eff;
  color: #409eff;
}

.page-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.page-info {
  color: #606266;
  font-size: 14px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
  font-size: 16px;
}
</style>
