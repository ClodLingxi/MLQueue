<script setup>
import { ref, computed, watch } from 'vue'
import {
  listTrainingQueues,
  createTrainingQueue,
  batchCreateTrainingQueues,
  deleteTrainingQueue,
  updateTrainingQueue,
  reorderTrainingQueues,
  getStatusColor,
  formatQueueStatus
} from '../../api/mlqueue-v2.js'

const props = defineProps({
  unitId: { type: String, required: true }
})
const emit = defineEmits(['back'])

const queues = ref([])
const loading = ref(false)
const showCreateDialog = ref(false)
const showBatchDialog = ref(false)
const draggedQueueId = ref(null)

// 参数表格形式
const parameterRows = ref([{ key: '', value: '' }])
const createForm = ref({ name: '' })
const batchForm = ref({ queues: '[]' })

// 将队列分为不可移动和可移动两组
const sortedQueues = computed(() => {
  const nonMovable = queues.value.filter(q =>
    q.status === 'running' || q.status === 'completed' || q.status === 'failed'
  )
  const movable = queues.value.filter(q => q.status === 'pending')

  // 按order排序 pending 队列（数字越小越先执行）
  movable.sort((a, b) => a.order - b.order)

  return {
    nonMovable,
    movable,
    all: [...nonMovable, ...movable]
  }
})

const fetchQueues = async () => {
  if (!props.unitId) return
  loading.value = true
  try {
    const response = await listTrainingQueues(props.unitId)
    queues.value = response.queues || []
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

// 打开创建对话框时加载默认参数
const openCreateDialog = () => {
  // 获取最新的已完成或运行中的队列参数
  const latestQueue = [...queues.value]
    .filter(q => q.status === 'completed' || q.status === 'running')
    .sort((a, b) => {
      const dateA = new Date(a.completed_at || a.started_at || a.created_at)
      const dateB = new Date(b.completed_at || b.started_at || b.created_at)
      return dateB - dateA
    })[0]

  if (latestQueue && latestQueue.parameters && Object.keys(latestQueue.parameters).length > 0) {
    // 将参数对象转换为表格行
    parameterRows.value = Object.entries(latestQueue.parameters).map(([key, value]) => ({
      key,
      value: String(value)
    }))
  } else {
    // 没有历史参数，使用空行
    parameterRows.value = [{ key: '', value: '' }]
  }

  createForm.value = { name: '' }
  showCreateDialog.value = true
}

const handleCreate = async () => {
  if (!createForm.value.name) {
    alert('请输入队列名称')
    return
  }

  // 将表格行转换为参数对象
  const parameters = {}
  parameterRows.value.forEach(row => {
    if (row.key.trim()) {
      // 尝试转换数值
      const trimmedValue = row.value.trim()
      if (trimmedValue === '') {
        parameters[row.key] = ''
      } else if (!isNaN(trimmedValue) && trimmedValue !== '') {
        parameters[row.key] = Number(trimmedValue)
      } else if (trimmedValue === 'true') {
        parameters[row.key] = true
      } else if (trimmedValue === 'false') {
        parameters[row.key] = false
      } else {
        parameters[row.key] = trimmedValue
      }
    }
  })

  try {
    await createTrainingQueue(props.unitId, {
      name: createForm.value.name,
      parameters
    })
    showCreateDialog.value = false
    createForm.value = { name: '' }
    parameterRows.value = [{ key: '', value: '' }]
    fetchQueues()
  } catch (err) {
    alert('创建失败: ' + err.message)
  }
}

// 添加参数行
const addParameterRow = () => {
  parameterRows.value.push({ key: '', value: '' })
}

// 删除参数行
const removeParameterRow = (index) => {
  if (parameterRows.value.length > 1) {
    parameterRows.value.splice(index, 1)
  }
}

const handleBatchCreate = async () => {
  try {
    const queuesData = JSON.parse(batchForm.value.queues)
    await batchCreateTrainingQueues(props.unitId, queuesData)
    showBatchDialog.value = false
    batchForm.value = { queues: '[]' }
    fetchQueues()
  } catch (err) {
    alert('批量创建失败: ' + err.message)
  }
}

const handleDelete = async (queueId) => {
  if (!confirm('确定删除？')) return
  try {
    await deleteTrainingQueue(queueId)
    fetchQueues()
  } catch (err) {
    alert('删除失败: ' + err.message)
  }
}

// ========== 拖拽功能 ==========

const handleDragStart = (event, queue) => {
  if (queue.status !== 'pending') {
    event.preventDefault()
    return
  }
  draggedQueueId.value = queue.queue_id
  event.dataTransfer.effectAllowed = 'move'
  event.dataTransfer.setData('text/html', event.target.innerHTML)
}

const handleDragOver = (event, queue) => {
  if (queue.status !== 'pending' || !draggedQueueId.value) {
    return
  }
  event.preventDefault()
  event.dataTransfer.dropEffect = 'move'
}

const handleDrop = async (event, targetQueue) => {
  event.preventDefault()

  if (!draggedQueueId.value || targetQueue.status !== 'pending') {
    draggedQueueId.value = null
    return
  }

  const draggedQueue = sortedQueues.value.movable.find(q => q.queue_id === draggedQueueId.value)
  if (!draggedQueue || draggedQueue.queue_id === targetQueue.queue_id) {
    draggedQueueId.value = null
    return
  }

  // 重新排序
  await reorderQueues(draggedQueue, targetQueue)
  draggedQueueId.value = null
}

const handleDragEnd = () => {
  draggedQueueId.value = null
}

// ========== 按键调整顺序 ==========

const moveUp = async (queue) => {
  const movableQueues = sortedQueues.value.movable
  const index = movableQueues.findIndex(q => q.queue_id === queue.queue_id)

  if (index <= 0) return // 已经在最前面

  const targetQueue = movableQueues[index - 1]
  await reorderQueues(queue, targetQueue)
}

const moveDown = async (queue) => {
  const movableQueues = sortedQueues.value.movable
  const index = movableQueues.findIndex(q => q.queue_id === queue.queue_id)

  if (index === -1 || index >= movableQueues.length - 1) return // 已经在最后面

  const targetQueue = movableQueues[index + 1]
  await reorderQueues(targetQueue, queue)
}

// ========== 重新排序逻辑 ==========

const reorderQueues = async (movedQueue, targetQueue) => {
  try {
    // 获取当前可移动队列列表
    const movableQueues = sortedQueues.value.movable

    // 找到两个队列在数组中的索引
    const movedIndex = movableQueues.findIndex(q => q.queue_id === movedQueue.queue_id)
    const targetIndex = movableQueues.findIndex(q => q.queue_id === targetQueue.queue_id)

    if (movedIndex === -1 || targetIndex === -1) return

    // 创建新的队列顺序数组
    const newOrder = [...movableQueues]

    // 移除被拖动的队列
    const [removed] = newOrder.splice(movedIndex, 1)

    // 插入到目标位置
    newOrder.splice(targetIndex, 0, removed)

    // 提取新顺序的队列ID数组
    const queueIds = newOrder.map(q => q.queue_id)

    // 调用 reorder API
    await reorderTrainingQueues(props.unitId, queueIds)

    // 重新获取队列列表
    await fetchQueues()
  } catch (err) {
    alert('调整顺序失败: ' + err.message)
  }
}

// 判断是否可以移动
const canMove = (queue) => {
  return queue.status === 'pending'
}

const canMoveUp = (queue) => {
  if (!canMove(queue)) return false
  const movableQueues = sortedQueues.value.movable
  const index = movableQueues.findIndex(q => q.queue_id === queue.queue_id)
  return index > 0
}

const canMoveDown = (queue) => {
  if (!canMove(queue)) return false
  const movableQueues = sortedQueues.value.movable
  const index = movableQueues.findIndex(q => q.queue_id === queue.queue_id)
  return index < movableQueues.length - 1
}

// 格式化 metric 值
const formatMetricValue = (value) => {
  if (typeof value === 'number') {
    // 如果是小数，保留4位
    return value % 1 !== 0 ? value.toFixed(4) : value
  }
  return value
}

// 格式化参数值
const formatParamValue = (value) => {
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

watch(() => props.unitId, () => {
  if (props.unitId) fetchQueues()
}, { immediate: true })
</script>

<template>
  <div class="queue-management">
    <div class="header">
      <button @click="emit('back')" class="btn-back">← 返回</button>
      <div class="actions">
        <button @click="fetchQueues" class="btn-refresh">刷新</button>
        <button @click="showBatchDialog = true" class="btn-secondary">批量创建</button>
        <button @click="openCreateDialog" class="btn-create">创建队列</button>
      </div>
    </div>

    <!-- 使用提示 -->
    <div class="tips">
      <p>💡 提示：只有"待执行"状态的队列可以拖动或调整顺序</p>
    </div>

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="queues.length > 0" class="queue-list">
      <!-- 不可移动的队列（running, completed, failed） -->
      <div v-if="sortedQueues.nonMovable.length > 0" class="queue-section">
        <h3 class="section-title">运行中 / 已完成</h3>
        <div
          v-for="queue in sortedQueues.nonMovable"
          :key="queue.queue_id"
          class="queue-item non-movable"
        >
          <div class="queue-content">
            <div class="queue-header">
              <div class="queue-title">
                <span class="drag-handle disabled">⋮⋮</span>
                <h4>{{ queue.name }}</h4>
              </div>
              <span class="status-badge" :style="{ backgroundColor: getStatusColor(queue.status) }">
                {{ formatQueueStatus(queue.status) }}
              </span>
            </div>
            <div class="queue-info">
              <span>执行顺序: {{ queue.order }}</span>
              <span>创建者: {{ queue.created_by }}</span>
              <span v-if="queue.started_at">开始时间: {{ new Date(queue.started_at).toLocaleString('zh-CN') }}</span>
              <span v-if="queue.completed_at">完成时间: {{ new Date(queue.completed_at).toLocaleString('zh-CN') }}</span>
            </div>
            <!-- 训练参数表格 -->
            <div class="queue-params">
              <strong>参数:</strong>
              <table class="params-table" v-if="queue.parameters && Object.keys(queue.parameters).length > 0">
                <tbody>
                  <tr v-for="(value, key) in queue.parameters" :key="key">
                    <td class="param-key">{{ key }}</td>
                    <td class="param-value">{{ formatParamValue(value) }}</td>
                  </tr>
                </tbody>
              </table>
              <p v-else class="no-params">无参数</p>
            </div>

            <!-- 训练结果 -->
            <div v-if="queue.result" class="queue-result">
              <strong>训练结果:</strong>
              <pre>{{ JSON.stringify(queue.result, null, 2) }}</pre>
            </div>

            <!-- 训练指标 Metrics -->
            <div v-if="queue.metrics && Object.keys(queue.metrics).length > 0" class="queue-metrics">
              <strong>训练指标 (Metrics):</strong>
              <div class="metrics-grid">
                <div v-for="(value, key) in queue.metrics" :key="key" class="metric-item">
                  <span class="metric-name">{{ key }}:</span>
                  <span class="metric-value">{{ formatMetricValue(value) }}</span>
                </div>
              </div>
            </div>

            <!-- 错误信息 -->
            <div v-if="queue.error_msg" class="queue-error">
              <strong>错误信息:</strong>
              <p>{{ queue.error_msg }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- 可移动的队列（pending） -->
      <div v-if="sortedQueues.movable.length > 0" class="queue-section">
        <h3 class="section-title">待执行队列（可拖动调整顺序）</h3>
        <div
          v-for="queue in sortedQueues.movable"
          :key="queue.queue_id"
          :class="['queue-item', 'movable', { 'dragging': draggedQueueId === queue.queue_id }]"
          draggable="true"
          @dragstart="handleDragStart($event, queue)"
          @dragover="handleDragOver($event, queue)"
          @drop="handleDrop($event, queue)"
          @dragend="handleDragEnd"
        >
          <div class="queue-content">
            <div class="queue-header">
              <div class="queue-title">
                <span class="drag-handle">⋮⋮</span>
                <h4>{{ queue.name }}</h4>
              </div>
              <span class="status-badge" :style="{ backgroundColor: getStatusColor(queue.status) }">
                {{ formatQueueStatus(queue.status) }}
              </span>
            </div>
            <div class="queue-info">
              <span>执行顺序: {{ queue.order }}</span>
              <span>创建者: {{ queue.created_by }}</span>
            </div>
            <!-- 训练参数表格 -->
            <div class="queue-params">
              <strong>参数:</strong>
              <table class="params-table" v-if="queue.parameters && Object.keys(queue.parameters).length > 0">
                <tbody>
                  <tr v-for="(value, key) in queue.parameters" :key="key">
                    <td class="param-key">{{ key }}</td>
                    <td class="param-value">{{ formatParamValue(value) }}</td>
                  </tr>
                </tbody>
              </table>
              <p v-else class="no-params">无参数</p>
            </div>

            <!-- 操作按钮 -->
            <div class="queue-actions">
              <div class="move-buttons">
                <button
                  @click="moveUp(queue)"
                  :disabled="!canMoveUp(queue)"
                  class="btn-move"
                  title="上移"
                >
                  ↑
                </button>
                <button
                  @click="moveDown(queue)"
                  :disabled="!canMoveDown(queue)"
                  class="btn-move"
                  title="下移"
                >
                  ↓
                </button>
              </div>
              <button @click="handleDelete(queue.queue_id)" class="btn-delete">删除</button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <p>暂无训练队列</p>
      <button @click="openCreateDialog" class="btn-primary">创建队列</button>
    </div>

    <!-- 创建对话框 -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <h3>创建训练队列</h3>
        <div class="form-group">
          <label>名称</label>
          <input v-model="createForm.name" placeholder="lr_0.001" />
        </div>
        <div class="form-group">
          <label>参数</label>
          <div class="params-editor">
            <table class="params-input-table">
              <thead>
                <tr>
                  <th>参数名</th>
                  <th>参数值</th>
                  <th width="60">操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, index) in parameterRows" :key="index">
                  <td>
                    <input
                      v-model="row.key"
                      type="text"
                      placeholder="learning_rate"
                      class="param-input"
                    />
                  </td>
                  <td>
                    <input
                      v-model="row.value"
                      type="text"
                      placeholder="0.001"
                      class="param-input"
                    />
                  </td>
                  <td class="action-cell">
                    <button
                      v-if="parameterRows.length > 1"
                      @click="removeParameterRow(index)"
                      class="btn-remove-row"
                      type="button"
                      title="删除"
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
            <button @click="addParameterRow" class="btn-add-row" type="button">
              + 添加参数
            </button>
          </div>
        </div>
        <div class="form-group">
          <p class="form-note">💡 执行顺序(order)由系统自动分配，新队列追加到末尾</p>
          <p class="form-note">💡 参数值会自动识别类型：数字、布尔值(true/false)或字符串</p>
        </div>
        <div class="dialog-actions">
          <button @click="showCreateDialog = false" class="btn-secondary">取消</button>
          <button @click="handleCreate" class="btn-primary">创建</button>
        </div>
      </div>
    </div>

    <!-- 批量创建对话框 -->
    <div v-if="showBatchDialog" class="dialog-overlay" @click.self="showBatchDialog = false">
      <div class="dialog">
        <h3>批量创建训练队列</h3>
        <div class="form-group">
          <label>队列数组 (JSON)</label>
          <textarea v-model="batchForm.queues" rows="10" placeholder='[{"name": "lr_0.001", "parameters": {"learning_rate": 0.001}}, {"name": "lr_0.01", "parameters": {"learning_rate": 0.01}}]'></textarea>
        </div>
        <div class="form-group">
          <p class="form-note">💡 队列将按数组顺序执行，系统会自动分配order值(0, 1, 2...)</p>
        </div>
        <div class="dialog-actions">
          <button @click="showBatchDialog = false" class="btn-secondary">取消</button>
          <button @click="handleBatchCreate" class="btn-primary">批量创建</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.queue-management {
  background: white;
  border-radius: 8px;
  padding: 20px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.actions {
  display: flex;
  gap: 10px;
}

.btn-back,
.btn-refresh,
.btn-create,
.btn-secondary {
  padding: 8px 16px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-back,
.btn-refresh {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-back:hover,
.btn-refresh:hover {
  background: #e4e7ed;
}

.btn-create {
  background: #409eff;
  color: white;
  border: none;
}

.btn-create:hover {
  background: #66b1ff;
}

.btn-secondary {
  background: white;
  color: #409eff;
  border: 1px solid #409eff;
}

.btn-secondary:hover {
  background: #ecf5ff;
}

/* 提示信息 */
.tips {
  background: #ecf5ff;
  border: 1px solid #d9ecff;
  border-radius: 4px;
  padding: 12px;
  margin-bottom: 20px;
}

.tips p {
  margin: 0;
  color: #409eff;
  font-size: 14px;
}

.loading {
  text-align: center;
  padding: 40px;
  color: #909399;
}

/* 队列列表 */
.queue-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.queue-section {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 15px;
  background: #fafafa;
}

.section-title {
  margin: 0 0 15px 0;
  font-size: 16px;
  color: #606266;
  padding-bottom: 10px;
  border-bottom: 2px solid #e4e7ed;
}

/* 队列项 */
.queue-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 15px;
  background: white;
  margin-bottom: 10px;
  transition: all 0.3s;
}

.queue-item:last-child {
  margin-bottom: 0;
}

/* 可移动队列 */
.queue-item.movable {
  cursor: move;
  border-left: 4px solid #409eff;
}

.queue-item.movable:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border-color: #66b1ff;
}

.queue-item.dragging {
  opacity: 0.5;
  border-color: #a0cfff;
}

/* 不可移动队列 */
.queue-item.non-movable {
  border-left: 4px solid #909399;
  background: #f5f7fa;
  cursor: not-allowed;
}

.queue-content {
  position: relative;
}

.queue-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
}

.queue-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 拖拽手柄 */
.drag-handle {
  font-size: 16px;
  color: #409eff;
  cursor: grab;
  user-select: none;
  line-height: 1;
}

.drag-handle.disabled {
  color: #c0c4cc;
  cursor: not-allowed;
}

.drag-handle:active {
  cursor: grabbing;
}

.queue-header h4 {
  margin: 0;
  font-size: 16px;
  color: #303133;
}

.status-badge {
  padding: 4px 12px;
  border-radius: 12px;
  color: white;
  font-size: 12px;
  white-space: nowrap;
}

.queue-info {
  display: flex;
  gap: 15px;
  font-size: 13px;
  color: #606266;
  margin-bottom: 10px;
}

.queue-params,
.queue-result,
.queue-metrics,
.queue-error {
  margin-bottom: 10px;
}

.queue-params strong,
.queue-result strong,
.queue-metrics strong,
.queue-error strong {
  display: block;
  margin-bottom: 5px;
  color: #606266;
  font-size: 13px;
}

.queue-result pre {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  margin: 0;
  font-family: 'Courier New', monospace;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.no-params {
  color: #909399;
  font-size: 12px;
  margin: 0;
  font-style: italic;
}

/* 参数表格样式 */
.params-table {
  width: 100%;
  border-collapse: collapse;
  background: #fafafa;
  border-radius: 4px;
  overflow: hidden;
  font-size: 13px;
}

.params-table tbody tr {
  border-bottom: 1px solid #e4e7ed;
}

.params-table tbody tr:last-child {
  border-bottom: none;
}

.params-table td {
  padding: 8px 12px;
}

.param-key {
  font-weight: 500;
  color: #606266;
  width: 40%;
  background: #f5f7fa;
}

.param-value {
  color: #303133;
  font-family: 'Courier New', monospace;
  word-break: break-word;
}

/* Metrics 网格显示 */
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 10px;
  background: #f0f9ff;
  padding: 12px;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.metric-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.metric-name {
  font-size: 12px;
  color: #606266;
  font-weight: 500;
}

.metric-value {
  font-size: 16px;
  color: #303133;
  font-weight: 600;
  font-family: 'Courier New', monospace;
}

/* 错误信息 */
.queue-error {
  background: #fef0f0;
  border-left: 3px solid #f56c6c;
  padding: 10px;
  border-radius: 4px;
}

.queue-error p {
  margin: 0;
  color: #f56c6c;
  font-size: 13px;
}

/* 操作按钮 */
.queue-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.move-buttons {
  display: flex;
  gap: 5px;
}

.btn-move {
  background: #409eff;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  font-weight: bold;
  transition: all 0.3s;
  min-width: 36px;
}

.btn-move:hover:not(:disabled) {
  background: #66b1ff;
  transform: translateY(-1px);
}

.btn-move:disabled {
  background: #c0c4cc;
  cursor: not-allowed;
  opacity: 0.5;
}

.btn-delete {
  background: #f56c6c;
  color: white;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-delete:hover {
  background: #f78989;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-state p {
  color: #909399;
  margin-bottom: 20px;
}

/* 对话框 */
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
  max-width: 600px;
  max-height: 80vh;
  overflow-y: auto;
}

.dialog h3 {
  margin: 0 0 20px 0;
  font-size: 18px;
  color: #303133;
}

.form-group {
  margin-bottom: 15px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.form-group input,
.form-group textarea {
  width: 100%;
  padding: 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-family: monospace;
  box-sizing: border-box;
  font-size: 14px;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #409eff;
}

.form-note {
  margin: 0 0 5px 0;
  padding: 8px 12px;
  background: #ecf5ff;
  border-left: 3px solid #409eff;
  color: #409eff;
  font-size: 12px;
  border-radius: 4px;
}

/* 参数编辑器 */
.params-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  background: #fafafa;
}

.params-input-table {
  width: 100%;
  border-collapse: collapse;
  margin-bottom: 10px;
}

.params-input-table thead th {
  text-align: left;
  padding: 8px;
  background: #f5f7fa;
  color: #606266;
  font-size: 13px;
  font-weight: 500;
  border-bottom: 2px solid #dcdfe6;
}

.params-input-table tbody td {
  padding: 6px 8px;
}

.param-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 13px;
  font-family: 'Courier New', monospace;
  box-sizing: border-box;
}

.param-input:focus {
  outline: none;
  border-color: #409eff;
}

.action-cell {
  text-align: center;
}

.btn-remove-row {
  background: #f56c6c;
  color: white;
  border: none;
  border-radius: 3px;
  width: 24px;
  height: 24px;
  cursor: pointer;
  font-size: 14px;
  line-height: 1;
  transition: all 0.3s;
}

.btn-remove-row:hover {
  background: #f78989;
}

.btn-add-row {
  background: #409eff;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 12px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.3s;
  width: 100%;
}

.btn-add-row:hover {
  background: #66b1ff;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.btn-primary {
  background: #409eff;
  color: white;
  border: none;
  padding: 8px 20px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-primary:hover {
  background: #66b1ff;
}
</style>