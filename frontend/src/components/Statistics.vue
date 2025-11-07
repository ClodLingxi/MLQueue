<script setup>
import { ref, onMounted } from 'vue'
import { getTaskStatistics } from '../api/mlqueue.js'

// 统计数据
const statistics = ref(null)
const loading = ref(false)
const error = ref(null)

// 日期范围
const startDate = ref('')
const endDate = ref('')

// 获取统计数据
const fetchStatistics = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await getTaskStatistics(
      startDate.value || undefined,
      endDate.value || undefined
    )
    statistics.value = response
  } catch (err) {
    error.value = err.message || '获取统计数据失败'
    console.error('获取统计数据失败:', err)
  } finally {
    loading.value = false
  }
}

// 设置默认日期范围(最近30天)
const setDefaultDateRange = () => {
  const end = new Date()
  const start = new Date()
  start.setDate(start.getDate() - 30)

  endDate.value = end.toISOString().split('T')[0]
  startDate.value = start.toISOString().split('T')[0]
}

// 应用筛选
const handleFilter = () => {
  fetchStatistics()
}

// 重置筛选
const handleReset = () => {
  setDefaultDateRange()
  fetchStatistics()
}

// 格式化日期
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleDateString('zh-CN')
}

// 组件挂载时加载数据
onMounted(() => {
  setDefaultDateRange()
  fetchStatistics()
})
</script>

<template>
  <div class="statistics">
    <h2>统计信息</h2>

    <!-- 日期筛选 -->
    <div class="filters">
      <div class="filter-group">
        <label>开始日期</label>
        <input v-model="startDate" type="date" />
      </div>
      <div class="filter-group">
        <label>结束日期</label>
        <input v-model="endDate" type="date" />
      </div>
      <div class="filter-actions">
        <button @click="handleFilter" class="btn-primary">查询</button>
        <button @click="handleReset" class="btn-secondary">重置</button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading && !statistics" class="loading">
      加载中...
    </div>

    <!-- 统计内容 -->
    <div v-else-if="statistics" class="content">
      <!-- 时间范围 -->
      <div class="period-info">
        <span>统计周期:</span>
        <strong>{{ formatDate(statistics.period.start) }}</strong>
        <span>至</span>
        <strong>{{ formatDate(statistics.period.end) }}</strong>
      </div>

      <!-- 关键指标卡片 -->
      <div class="metrics-grid">
        <div class="metric-card card-blue">
          <div class="metric-icon">📊</div>
          <div class="metric-content">
            <div class="metric-label">总任务数</div>
            <div class="metric-value">{{ statistics.statistics.total_tasks }}</div>
          </div>
        </div>

        <div class="metric-card card-green">
          <div class="metric-icon">✅</div>
          <div class="metric-content">
            <div class="metric-label">完成任务</div>
            <div class="metric-value">{{ statistics.statistics.completed_tasks }}</div>
          </div>
        </div>

        <div class="metric-card card-red">
          <div class="metric-icon">❌</div>
          <div class="metric-content">
            <div class="metric-label">失败任务</div>
            <div class="metric-value">{{ statistics.statistics.failed_tasks }}</div>
          </div>
        </div>

        <div class="metric-card card-purple">
          <div class="metric-icon">⏱️</div>
          <div class="metric-content">
            <div class="metric-label">平均耗时</div>
            <div class="metric-value">{{ statistics.statistics.average_duration }}</div>
          </div>
        </div>
      </div>

      <!-- 成功率 -->
      <div class="success-rate-section">
        <h3>任务成功率</h3>
        <div class="success-rate-container">
          <div class="rate-circle">
            <svg viewBox="0 0 100 100" class="circular-chart">
              <circle class="circle-bg" cx="50" cy="50" r="40" />
              <circle
                class="circle-progress"
                cx="50"
                cy="50"
                r="40"
                :style="{
                  strokeDashoffset: 251.2 * (1 - statistics.statistics.success_rate)
                }"
              />
            </svg>
            <div class="rate-text">
              {{ Math.round(statistics.statistics.success_rate * 100) }}%
            </div>
          </div>
          <div class="rate-details">
            <div class="rate-item">
              <span class="rate-label">成功率:</span>
              <span class="rate-value success">
                {{ (statistics.statistics.success_rate * 100).toFixed(2) }}%
              </span>
            </div>
            <div class="rate-item">
              <span class="rate-label">完成数:</span>
              <span class="rate-value">{{ statistics.statistics.completed_tasks }}</span>
            </div>
            <div class="rate-item">
              <span class="rate-label">失败数:</span>
              <span class="rate-value failed">{{ statistics.statistics.failed_tasks }}</span>
            </div>
            <div class="rate-item">
              <span class="rate-label">总数:</span>
              <span class="rate-value">{{ statistics.statistics.total_tasks }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 详细统计 -->
      <div class="detailed-stats">
        <h3>详细统计</h3>
        <div class="stats-table">
          <div class="stats-row">
            <div class="stats-label">总任务数</div>
            <div class="stats-value">{{ statistics.statistics.total_tasks }}</div>
          </div>
          <div class="stats-row">
            <div class="stats-label">已完成任务</div>
            <div class="stats-value text-success">{{ statistics.statistics.completed_tasks }}</div>
          </div>
          <div class="stats-row">
            <div class="stats-label">失败任务</div>
            <div class="stats-value text-danger">{{ statistics.statistics.failed_tasks }}</div>
          </div>
          <div class="stats-row">
            <div class="stats-label">平均执行时长</div>
            <div class="stats-value">{{ statistics.statistics.average_duration }}</div>
          </div>
          <div class="stats-row">
            <div class="stats-label">成功率</div>
            <div class="stats-value text-success">
              {{ (statistics.statistics.success_rate * 100).toFixed(2) }}%
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.statistics {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.statistics h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: #303133;
}

.statistics h3 {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #606266;
}

.filters {
  display: flex;
  gap: 15px;
  align-items: flex-end;
  margin-bottom: 30px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.filter-group {
  flex: 1;
}

.filter-group label {
  display: block;
  margin-bottom: 8px;
  color: #606266;
  font-size: 14px;
}

.filter-group input {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.filter-group input:focus {
  outline: none;
  border-color: #409eff;
}

.filter-actions {
  display: flex;
  gap: 10px;
}

.btn-primary,
.btn-secondary {
  padding: 8px 20px;
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
  background: white;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-secondary:hover {
  background: #f5f7fa;
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

.period-info {
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 20px;
  font-size: 14px;
  color: #606266;
}

.period-info span {
  margin: 0 8px;
}

.period-info strong {
  color: #303133;
}

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.metric-card {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 20px;
  border-radius: 8px;
  color: white;
}

.card-blue {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.card-green {
  background: linear-gradient(135deg, #84fab0 0%, #8fd3f4 100%);
}

.card-red {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
}

.card-purple {
  background: linear-gradient(135deg, #a8edea 0%, #fed6e3 100%);
}

.metric-icon {
  font-size: 40px;
}

.metric-content {
  flex: 1;
}

.metric-label {
  font-size: 14px;
  opacity: 0.9;
  margin-bottom: 5px;
}

.metric-value {
  font-size: 32px;
  font-weight: 600;
}

.success-rate-section {
  margin-bottom: 30px;
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.success-rate-container {
  display: flex;
  align-items: center;
  gap: 40px;
  margin-top: 20px;
}

.rate-circle {
  position: relative;
  width: 150px;
  height: 150px;
}

.circular-chart {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}

.circle-bg {
  fill: none;
  stroke: #e4e7ed;
  stroke-width: 8;
}

.circle-progress {
  fill: none;
  stroke: #67c23a;
  stroke-width: 8;
  stroke-linecap: round;
  stroke-dasharray: 251.2;
  transition: stroke-dashoffset 0.5s;
}

.rate-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 32px;
  font-weight: 600;
  color: #303133;
}

.rate-details {
  flex: 1;
}

.rate-item {
  display: flex;
  justify-content: space-between;
  padding: 10px 0;
  border-bottom: 1px solid #e4e7ed;
}

.rate-item:last-child {
  border-bottom: none;
}

.rate-label {
  color: #606266;
  font-size: 14px;
}

.rate-value {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.rate-value.success {
  color: #67c23a;
}

.rate-value.failed {
  color: #f56c6c;
}

.detailed-stats {
  padding: 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.stats-table {
  background: white;
  border-radius: 4px;
  overflow: hidden;
}

.stats-row {
  display: flex;
  justify-content: space-between;
  padding: 15px 20px;
  border-bottom: 1px solid #e4e7ed;
}

.stats-row:last-child {
  border-bottom: none;
}

.stats-label {
  color: #606266;
  font-size: 14px;
}

.stats-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.text-success {
  color: #67c23a;
}

.text-danger {
  color: #f56c6c;
}
</style>
