<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { getQueueStatus, pauseQueue, resumeQueue } from '../api/mlqueue.js'

// 队列数据
const queueData = ref(null)
const loading = ref(false)
const error = ref(null)

// 自动刷新定时器
let refreshInterval = null

// 获取队列状态
const fetchQueueStatus = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getQueueStatus()
    queueData.value = response
  } catch (err) {
    error.value = err.message || '获取队列状态失败'
    console.error('获取队列状态失败:', err)
  } finally {
    loading.value = false
  }
}

// 暂停队列
const handlePause = async () => {
  if (!confirm('确定要暂停队列吗?')) return

  try {
    await pauseQueue()
    await fetchQueueStatus()
  } catch (err) {
    alert('暂停队列失败: ' + err.message)
  }
}

// 恢复队列
const handleResume = async () => {
  try {
    await resumeQueue()
    await fetchQueueStatus()
  } catch (err) {
    alert('恢复队列失败: ' + err.message)
  }
}

// 刷新
const handleRefresh = () => {
  fetchQueueStatus()
}

// 格式化时间
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

// 计算统计百分比
const getPercentage = (value, total) => {
  if (!total) return 0
  return Math.round((value / total) * 100)
}

// 组件挂载时开始自动刷新
onMounted(() => {
  fetchQueueStatus()

  // 每 5 秒自动刷新
  refreshInterval = setInterval(() => {
    fetchQueueStatus()
  }, 5000)
})

// 组件卸载时清除定时器
onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<template>
  <div class="queue-status">
    <div class="header">
      <h2>队列状态</h2>
      <div class="actions">
        <button @click="handleRefresh" class="btn-refresh">刷新</button>
        <button
          v-if="queueData?.queue_status !== 'paused'"
          @click="handlePause"
          class="btn-warning"
        >
          暂停队列
        </button>
        <button
          v-else
          @click="handleResume"
          class="btn-success"
        >
          恢复队列
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading && !queueData" class="loading">
      加载中...
    </div>

    <!-- 队列信息 -->
    <div v-else-if="queueData" class="content">
      <!-- 基本信息 -->
      <div class="info-cards">
        <div class="info-card">
          <div class="card-label">队列名称</div>
          <div class="card-value">{{ queueData.queue_name }}</div>
        </div>
        <div class="info-card">
          <div class="card-label">队列长度</div>
          <div class="card-value highlight">{{ queueData.queue_length }}</div>
        </div>
        <div class="info-card">
          <div class="card-label">预计等待时间</div>
          <div class="card-value">{{ queueData.estimated_wait_time }}</div>
        </div>
      </div>

      <!-- 任务统计 -->
      <div class="section">
        <h3>任务统计</h3>
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon" style="background: #909399;">
              <span class="icon">⏳</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">待处理</div>
              <div class="stat-value">{{ queueData.statistics.pending }}</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon" style="background: #e6a23c;">
              <span class="icon">📋</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">队列中</div>
              <div class="stat-value">{{ queueData.statistics.queued }}</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon" style="background: #409eff;">
              <span class="icon">▶️</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">运行中</div>
              <div class="stat-value">{{ queueData.statistics.running }}</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon" style="background: #67c23a;">
              <span class="icon">✅</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">已完成</div>
              <div class="stat-value">{{ queueData.statistics.completed }}</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon" style="background: #f56c6c;">
              <span class="icon">❌</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">失败</div>
              <div class="stat-value">{{ queueData.statistics.failed }}</div>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon" style="background: #909399;">
              <span class="icon">🚫</span>
            </div>
            <div class="stat-info">
              <div class="stat-label">已取消</div>
              <div class="stat-value">{{ queueData.statistics.cancelled }}</div>
            </div>
          </div>
        </div>

        <!-- 统计图表 -->
        <div class="stats-chart">
          <div class="chart-bar">
            <div
              class="bar-segment"
              style="background: #67c23a;"
              :style="{ width: getPercentage(queueData.statistics.completed,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) + '%' }"
            >
              <span v-if="getPercentage(queueData.statistics.completed,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) > 10">
                {{ getPercentage(queueData.statistics.completed,
                  queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) }}%
              </span>
            </div>
            <div
              class="bar-segment"
              style="background: #f56c6c;"
              :style="{ width: getPercentage(queueData.statistics.failed,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) + '%' }"
            >
              <span v-if="getPercentage(queueData.statistics.failed,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) > 10">
                {{ getPercentage(queueData.statistics.failed,
                  queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) }}%
              </span>
            </div>
            <div
              class="bar-segment"
              style="background: #909399;"
              :style="{ width: getPercentage(queueData.statistics.cancelled,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) + '%' }"
            >
              <span v-if="getPercentage(queueData.statistics.cancelled,
                queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) > 10">
                {{ getPercentage(queueData.statistics.cancelled,
                  queueData.statistics.completed + queueData.statistics.failed + queueData.statistics.cancelled) }}%
              </span>
            </div>
          </div>
          <div class="chart-legend">
            <span><span class="legend-dot" style="background: #67c23a;"></span> 已完成</span>
            <span><span class="legend-dot" style="background: #f56c6c;"></span> 失败</span>
            <span><span class="legend-dot" style="background: #909399;"></span> 已取消</span>
          </div>
        </div>
      </div>

      <!-- 当前运行的任务 -->
      <div class="section">
        <h3>当前运行的任务</h3>
        <div v-if="queueData.current_tasks && queueData.current_tasks.length > 0" class="current-tasks">
          <div v-for="task in queueData.current_tasks" :key="task.task_id" class="task-card">
            <div class="task-header">
              <code>{{ task.task_id }}</code>
              <span class="status-running">运行中</span>
            </div>
            <div class="task-name">{{ task.name }}</div>
            <div class="task-time">开始时间: {{ formatDate(task.started_at) }}</div>
          </div>
        </div>
        <div v-else class="empty-state">
          当前没有运行中的任务
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.queue-status {
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

.btn-refresh,
.btn-warning,
.btn-success {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-refresh {
  background: #409eff;
  color: white;
}

.btn-refresh:hover {
  background: #66b1ff;
}

.btn-warning {
  background: #e6a23c;
  color: white;
}

.btn-warning:hover {
  background: #ebb563;
}

.btn-success {
  background: #67c23a;
  color: white;
}

.btn-success:hover {
  background: #85ce61;
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

.info-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.info-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
  border-radius: 8px;
  color: white;
}

.card-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 8px;
}

.card-value {
  font-size: 28px;
  font-weight: 600;
}

.card-value.highlight {
  color: #ffd93d;
}

.section {
  margin-bottom: 30px;
}

.section h3 {
  margin: 0 0 20px 0;
  font-size: 16px;
  color: #606266;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 15px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stat-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.stats-chart {
  margin-top: 20px;
}

.chart-bar {
  display: flex;
  height: 40px;
  border-radius: 20px;
  overflow: hidden;
  margin-bottom: 10px;
}

.bar-segment {
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 12px;
  font-weight: 600;
  transition: width 0.3s;
}

.chart-legend {
  display: flex;
  gap: 20px;
  justify-content: center;
  font-size: 14px;
  color: #606266;
}

.legend-dot {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  margin-right: 5px;
}

.current-tasks {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 15px;
}

.task-card {
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  padding: 15px;
  transition: all 0.3s;
}

.task-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.task-header code {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  font-family: monospace;
}

.status-running {
  background: #409eff;
  color: white;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
}

.task-name {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.task-time {
  font-size: 12px;
  color: #909399;
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: #909399;
}
</style>
