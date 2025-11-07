<script setup>
import { ref, onMounted, watch } from 'vue'
import { getTask, getTaskLogs, cancelTask, uploadTaskResult } from '../api/mlqueue.js'

const props = defineProps({
  taskId: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['back'])

// 任务数据
const task = ref(null)
const logs = ref([])
const loading = ref(false)
const error = ref(null)

// 日志选项
const showLogs = ref(false)
const logLines = ref(100)

// 上传结果表单
const showResultForm = ref(false)
const resultData = ref({
  final_loss: '',
  accuracy: '',
  precision: '',
  recall: ''
})

// 状态颜色映射
const statusColors = {
  pending: '#909399',
  queued: '#e6a23c',
  running: '#409eff',
  completed: '#67c23a',
  failed: '#f56c6c',
  cancelled: '#909399'
}

const statusLabels = {
  pending: '待处理',
  queued: '队列中',
  running: '运行中',
  completed: '已完成',
  failed: '失败',
  cancelled: '已取消'
}

// 获取任务详情
const fetchTask = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getTask(props.taskId)
    task.value = response
  } catch (err) {
    error.value = err.message || '获取任务详情失败'
    console.error('获取任务详情失败:', err)
  } finally {
    loading.value = false
  }
}

// 获取任务日志
const fetchLogs = async () => {
  try {
    const response = await getTaskLogs(props.taskId, logLines.value)
    logs.value = response.logs || []
  } catch (err) {
    console.error('获取日志失败:', err)
  }
}

// 取消任务
const handleCancel = async () => {
  if (!confirm('确定要取消这个任务吗?')) return

  try {
    await cancelTask(props.taskId, '用户手动取消')
    await fetchTask()
  } catch (err) {
    alert('取消任务失败: ' + err.message)
  }
}

// 上传结果
const handleUploadResult = async () => {
  try {
    const result = {
      final_loss: parseFloat(resultData.value.final_loss),
      accuracy: parseFloat(resultData.value.accuracy),
      metrics: {
        precision: parseFloat(resultData.value.precision),
        recall: parseFloat(resultData.value.recall)
      }
    }

    await uploadTaskResult(props.taskId, result)
    showResultForm.value = false
    await fetchTask()
    alert('结果上传成功')
  } catch (err) {
    alert('上传结果失败: ' + err.message)
  }
}

// 切换日志显示
const toggleLogs = () => {
  showLogs.value = !showLogs.value
  if (showLogs.value) {
    fetchLogs()
  }
}

// 刷新
const handleRefresh = () => {
  fetchTask()
  if (showLogs.value) {
    fetchLogs()
  }
}

// 格式化时间
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 格式化 JSON
const formatJson = (obj) => {
  return JSON.stringify(obj, null, 2)
}

// 监听 taskId 变化
watch(() => props.taskId, () => {
  fetchTask()
}, { immediate: true })

onMounted(() => {
  // 如果任务正在运行,设置自动刷新
  const interval = setInterval(() => {
    if (task.value && (task.value.status === 'running' || task.value.status === 'queued')) {
      fetchTask()
      if (showLogs.value) {
        fetchLogs()
      }
    }
  }, 5000)

  // 组件卸载时清除定时器
  return () => clearInterval(interval)
})
</script>

<template>
  <div class="task-detail">
    <div class="header">
      <button @click="emit('back')" class="btn-back">← 返回列表</button>
      <button @click="handleRefresh" class="btn-refresh">刷新</button>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading && !task" class="loading">
      加载中...
    </div>

    <!-- 任务详情 -->
    <div v-else-if="task" class="content">
      <div class="section">
        <div class="section-header">
          <h2>{{ task.name }}</h2>
          <span
            class="status-badge"
            :style="{ backgroundColor: statusColors[task.status] }"
          >
            {{ statusLabels[task.status] }}
          </span>
        </div>

        <div class="info-grid">
          <div class="info-item">
            <label>任务 ID</label>
            <code>{{ task.task_id }}</code>
          </div>
          <div class="info-item">
            <label>优先级</label>
            <span>{{ task.priority }}</span>
          </div>
          <div class="info-item">
            <label>创建时间</label>
            <span>{{ formatDate(task.created_at) }}</span>
          </div>
          <div class="info-item">
            <label>开始时间</label>
            <span>{{ formatDate(task.started_at) }}</span>
          </div>
          <div class="info-item">
            <label>完成时间</label>
            <span>{{ formatDate(task.completed_at) }}</span>
          </div>
        </div>
      </div>

      <!-- 配置信息 -->
      <div class="section">
        <h3>训练配置</h3>
        <pre class="json-display">{{ formatJson(task.config) }}</pre>
      </div>

      <!-- 训练结果 -->
      <div v-if="task.result" class="section">
        <h3>训练结果</h3>
        <pre class="json-display">{{ formatJson(task.result) }}</pre>
      </div>

      <!-- 错误信息 -->
      <div v-if="task.error_message" class="section">
        <h3>错误信息</h3>
        <div class="error-box">{{ task.error_message }}</div>
      </div>

      <!-- 操作按钮 -->
      <div class="actions">
        <button
          v-if="task.status === 'pending' || task.status === 'queued' || task.status === 'running'"
          @click="handleCancel"
          class="btn-danger"
        >
          取消任务
        </button>
        <button
          v-if="task.status === 'running'"
          @click="showResultForm = !showResultForm"
          class="btn-primary"
        >
          上传结果
        </button>
        <button @click="toggleLogs" class="btn-secondary">
          {{ showLogs ? '隐藏日志' : '查看日志' }}
        </button>
      </div>

      <!-- 上传结果表单 -->
      <div v-if="showResultForm" class="section">
        <h3>上传训练结果</h3>
        <div class="result-form">
          <div class="form-row">
            <div class="form-group">
              <label>Final Loss</label>
              <input v-model="resultData.final_loss" type="number" step="0.001" />
            </div>
            <div class="form-group">
              <label>Accuracy</label>
              <input v-model="resultData.accuracy" type="number" step="0.01" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Precision</label>
              <input v-model="resultData.precision" type="number" step="0.01" />
            </div>
            <div class="form-group">
              <label>Recall</label>
              <input v-model="resultData.recall" type="number" step="0.01" />
            </div>
          </div>
          <button @click="handleUploadResult" class="btn-primary">提交结果</button>
        </div>
      </div>

      <!-- 日志显示 -->
      <div v-if="showLogs" class="section">
        <div class="log-header">
          <h3>任务日志</h3>
          <div class="log-controls">
            <label>显示行数:</label>
            <input v-model.number="logLines" type="number" min="10" max="1000" />
            <button @click="fetchLogs" class="btn-secondary">刷新日志</button>
          </div>
        </div>
        <div class="log-display">
          <div v-for="(log, index) in logs" :key="index" class="log-line">
            <span class="log-time">{{ formatDate(log.timestamp) }}</span>
            <span :class="['log-level', `level-${log.level.toLowerCase()}`]">
              {{ log.level }}
            </span>
            <span class="log-message">{{ log.message }}</span>
          </div>
          <div v-if="logs.length === 0" class="log-empty">
            暂无日志
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.task-detail {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
}

.btn-back,
.btn-refresh {
  padding: 8px 16px;
  border: 1px solid #dcdfe6;
  background: white;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-back:hover,
.btn-refresh:hover {
  border-color: #409eff;
  color: #409eff;
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

.section {
  margin-bottom: 30px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-bottom: 20px;
}

.section-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.section h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #606266;
}

.status-badge {
  display: inline-block;
  padding: 6px 16px;
  border-radius: 14px;
  color: white;
  font-size: 14px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
}

.info-item label {
  display: block;
  color: #909399;
  font-size: 12px;
  margin-bottom: 4px;
}

.info-item span,
.info-item code {
  display: block;
  color: #303133;
  font-size: 14px;
}

.info-item code {
  background: #f5f7fa;
  padding: 4px 8px;
  border-radius: 3px;
  font-family: monospace;
}

.json-display {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 13px;
  font-family: 'Courier New', monospace;
  color: #303133;
}

.error-box {
  background: #fef0f0;
  color: #f56c6c;
  padding: 15px;
  border-radius: 4px;
  border-left: 4px solid #f56c6c;
}

.actions {
  display: flex;
  gap: 10px;
  margin: 20px 0;
}

.btn-primary,
.btn-secondary,
.btn-danger {
  padding: 10px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-primary {
  background: #409eff;
  color: white;
}

.btn-primary:hover {
  background: #66b1ff;
}

.btn-secondary {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-secondary:hover {
  background: #e4e7ed;
}

.btn-danger {
  background: #f56c6c;
  color: white;
}

.btn-danger:hover {
  background: #f78989;
}

.result-form {
  background: #f5f7fa;
  padding: 20px;
  border-radius: 4px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  color: #606266;
  font-size: 14px;
}

.form-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.log-controls {
  display: flex;
  align-items: center;
  gap: 10px;
}

.log-controls label {
  font-size: 14px;
  color: #606266;
}

.log-controls input {
  width: 80px;
  padding: 6px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
}

.log-display {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 15px;
  border-radius: 4px;
  max-height: 500px;
  overflow-y: auto;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.log-line {
  margin-bottom: 5px;
  line-height: 1.6;
}

.log-time {
  color: #6a9955;
  margin-right: 10px;
}

.log-level {
  display: inline-block;
  width: 60px;
  margin-right: 10px;
  font-weight: bold;
}

.level-info {
  color: #4fc3f7;
}

.level-warn {
  color: #ffb74d;
}

.level-error {
  color: #e57373;
}

.log-message {
  color: #d4d4d4;
}

.log-empty {
  text-align: center;
  color: #909399;
  padding: 20px;
}
</style>
