<script setup>
import { ref, onMounted, watch } from 'vue'
import {
  listTrainingUnits,
  createTrainingUnit,
  updateTrainingUnit,
  deleteTrainingUnit,
  getTrainingUnit
} from '@/api/mlqueue-v2.js'
import { getStatusColor } from '@/api/mlqueue-v2.js'

const props = defineProps({
  groupId: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['select-unit', 'back'])

const units = ref([])
const loading = ref(false)
const error = ref(null)
const showCreateDialog = ref(false)

const createForm = ref({
  name: '',
  description: '',
  config: '{}'
})

const fetchUnits = async () => {
  if (!props.groupId) return
  loading.value = true
  error.value = null
  try {
    const response = await listTrainingUnits(props.groupId)
    units.value = response.units || []
  } catch (err) {
    error.value = err.message || '获取训练单元失败'
  } finally {
    loading.value = false
  }
}

const handleCreate = async () => {
  if (!createForm.value.name) {
    alert('请输入训练单元名称')
    return
  }
  try {
    let config = {}
    if (createForm.value.config.trim()) {
      config = JSON.parse(createForm.value.config)
    }
    await createTrainingUnit(props.groupId, {
      name: createForm.value.name,
      description: createForm.value.description,
      config
    })
    showCreateDialog.value = false
    createForm.value = { name: '', description: '', config: '{}' }
    await fetchUnits()
  } catch (err) {
    alert('创建失败: ' + err.message)
  }
}

const handleDelete = async (unitId) => {
  if (!confirm('确定删除此训练单元？')) return
  try {
    await deleteTrainingUnit(unitId)
    await fetchUnits()
  } catch (err) {
    alert('删除失败: ' + err.message)
  }
}

// 格式化连接状态
const formatConnectionStatus = (status) => {
  const statusMap = {
    connected: { text: '已连接', icon: '🟢', color: '#67c23a' },
    disconnected: { text: '未连接', icon: '🔴', color: '#909399' }
  }
  return statusMap[status] || statusMap.disconnected
}

// 格式化心跳时间
const formatHeartbeat = (timestamp) => {
  if (!timestamp) return '从未连接'
  const date = new Date(timestamp)
  const now = new Date()
  const diff = Math.floor((now - date) / 1000) // 秒

  if (diff < 10) return '刚刚'
  if (diff < 60) return `${diff}秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  return date.toLocaleString('zh-CN')
}

watch(() => props.groupId, () => {
  if (props.groupId) fetchUnits()
}, { immediate: true })
</script>

<template>
  <div class="unit-management">
    <div class="header">
      <button @click="emit('back')" class="btn-back">← 返回组列表</button>
      <div class="actions">
        <button @click="fetchUnits" class="btn-refresh">刷新</button>
        <button @click="showCreateDialog = true" class="btn-create">创建训练单元</button>
      </div>
    </div>

    <div v-if="error" class="error-message">{{ error }}</div>
    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="units.length > 0" class="units-grid">
      <div
          v-for="unit in units"
          :key="unit.unit_id"
          class="unit-card"
          @click="emit('select-unit', unit)"
      >
        <h3>{{ unit.name }}</h3>

        <p class="unit-description">{{ unit.description || '暂无描述' }}</p>

        <div class="unit-info">
          <div class="info-row">
            <span class="info-label">状态:</span>
            <span
                class="status-badge"
                :style="{ backgroundColor: getStatusColor(unit.status) }"
            >
              {{ unit.status || 'idle' }}
            </span>
          </div>
          <div class="info-row">
            <span class="info-label">队列数:</span>
            <span>{{ unit.queue_count || 0 }}</span>
          </div>
          <div class="info-row">
            <span class="info-label">版本:</span>
            <span>v{{ unit.version }}</span>
          </div>
          <div class="info-row" v-if="unit.connection_status === 'connected'">
            <span class="info-label">最后心跳:</span>
            <span class="heartbeat-time">{{ formatHeartbeat(unit.last_heartbeat) }}</span>
          </div>
        </div>

        <!-- Python 连接状态 - 右下角 -->
        <div class="connection-indicator">
          <span class="connection-icon">{{ formatConnectionStatus(unit.connection_status).icon }}</span>
          <span
            class="connection-text"
            :style="{ color: formatConnectionStatus(unit.connection_status).color }"
          >
            {{ formatConnectionStatus(unit.connection_status).text }}
          </span>
        </div>

        <div class="card-actions" @click.stop>
          <button @click="handleDelete(unit.unit_id)" class="btn-icon">🗑️</button>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <p>暂无训练单元</p>
      <button @click="showCreateDialog = true" class="btn-primary">创建训练单元</button>
    </div>

    <!-- 创建对话框 -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <h3>创建训练单元</h3>
        <div class="form-group">
          <label>名称 *</label>
          <input v-model="createForm.name" type="text" placeholder="CNN超参数搜索" />
        </div>
        <div class="form-group">
          <label>描述</label>
          <textarea v-model="createForm.description" rows="2" placeholder="实验描述"></textarea>
        </div>
        <div class="form-group">
          <label>配置 (JSON)</label>
          <textarea v-model="createForm.config" rows="4" placeholder='{"model": "CNN"}'></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="showCreateDialog = false" class="btn-secondary">取消</button>
          <button @click="handleCreate" class="btn-primary">创建</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.unit-management {
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

.btn-back {
  padding: 8px 16px;
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.actions {
  display: flex;
  gap: 10px;
}

.btn-refresh,
.btn-create {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-refresh {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-create {
  background: #409eff;
  color: white;
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

.units-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.unit-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  position: relative;
  transition: all 0.3s;
}

.unit-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 8px rgba(64, 158, 255, 0.1);
}

.unit-card h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #303133;
}

/* Python 连接状态指示器 - 右下角 */
.connection-indicator {
  position: absolute;
  bottom: 15px;
  right: 15px;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  background: #f5f7fa;
  border-radius: 12px;
  font-size: 12px;
  border: 1px solid #e4e7ed;
}

.connection-icon {
  font-size: 10px;
  line-height: 1;
}

.connection-text {
  font-weight: 500;
  line-height: 1;
}

.unit-description {
  color: #606266;
  font-size: 14px;
  margin: 0 0 12px 0;
  min-height: 20px;
}

.unit-info {
  display: flex;
  flex-direction: column;
  gap: 6px;
  font-size: 13px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-label {
  color: #909399;
  min-width: 60px;
}

.heartbeat-time {
  color: #67c23a;
  font-size: 12px;
  font-style: italic;
}

.status-badge {
  padding: 2px 8px;
  border-radius: 10px;
  color: white;
  font-size: 11px;
}

.card-actions {
  position: absolute;
  top: 15px;
  right: 15px;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  opacity: 0.6;
}

.btn-icon:hover {
  opacity: 1;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-state p {
  color: #909399;
  margin-bottom: 20px;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.dialog {
  background: white;
  border-radius: 8px;
  padding: 24px;
  width: 90%;
  max-width: 500px;
}

.dialog h3 {
  margin: 0 0 20px 0;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-size: 14px;
  color: #606266;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
  font-family: monospace;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.btn-primary,
.btn-secondary {
  padding: 8px 20px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: #409eff;
  color: white;
}

.btn-secondary {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}
</style>
