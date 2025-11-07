<script setup>
import { ref, onMounted } from 'vue'
import { createTask, batchCreateTasks, getConfigTemplates } from '../api/mlqueue.js'

const emit = defineEmits(['task-created'])

// 表单数据
const taskName = ref('')
const taskPriority = ref(0)
const taskDescription = ref('')

// 配置参数
const hiddenSize = ref(64)
const learningRate = ref(0.001)
const epochs = ref(100)
const layers = ref(2)

// 批量创建
const batchMode = ref(false)
const batchConfig = ref('')

// 配置模板
const templates = ref([])
const selectedTemplate = ref('')

// 状态
const loading = ref(false)
const error = ref(null)
const success = ref(false)

// 获取配置模板
const fetchTemplates = async () => {
  try {
    const response = await getConfigTemplates()
    templates.value = response.templates || []
  } catch (err) {
    console.error('获取配置模板失败:', err)
  }
}

// 应用模板
const applyTemplate = () => {
  if (!selectedTemplate.value) return

  const template = templates.value.find(t => t.name === selectedTemplate.value)
  if (template && template.config) {
    hiddenSize.value = template.config.hidden_size || 64
    learningRate.value = template.config.learning_rate || 0.001
    epochs.value = template.config.epochs || 100
    layers.value = template.config.layers || 2
  }
}

// 创建单个任务
const handleCreateTask = async () => {
  if (!taskName.value) {
    alert('请输入任务名称')
    return
  }

  loading.value = true
  error.value = null
  success.value = false

  try {
    const taskData = {
      name: taskName.value,
      config: {
        hidden_size: hiddenSize.value,
        learning_rate: learningRate.value,
        epochs: epochs.value,
        layers: layers.value
      },
      priority: taskPriority.value,
      metadata: {
        description: taskDescription.value
      }
    }

    await createTask(taskData)
    success.value = true

    // 重置表单
    setTimeout(() => {
      emit('task-created')
    }, 1500)
  } catch (err) {
    error.value = err.message || '创建任务失败'
    console.error('创建任务失败:', err)
  } finally {
    loading.value = false
  }
}

// 批量创建任务
const handleBatchCreate = async () => {
  if (!batchConfig.value) {
    alert('请输入批量配置')
    return
  }

  loading.value = true
  error.value = null
  success.value = false

  try {
    const tasks = JSON.parse(batchConfig.value)
    if (!Array.isArray(tasks)) {
      throw new Error('配置必须是一个数组')
    }

    await batchCreateTasks(tasks)
    success.value = true

    setTimeout(() => {
      emit('task-created')
    }, 1500)
  } catch (err) {
    if (err instanceof SyntaxError) {
      error.value = 'JSON 格式错误: ' + err.message
    } else {
      error.value = err.message || '批量创建任务失败'
    }
    console.error('批量创建任务失败:', err)
  } finally {
    loading.value = false
  }
}

// 填充示例配置
const fillExample = () => {
  batchConfig.value = JSON.stringify([
    {
      name: "小模型训练",
      config: { hidden_size: 64, learning_rate: 0.001, epochs: 50 },
      priority: 0
    },
    {
      name: "中模型训练",
      config: { hidden_size: 128, learning_rate: 0.001, epochs: 100 },
      priority: 1
    },
    {
      name: "大模型训练",
      config: { hidden_size: 256, learning_rate: 0.0005, epochs: 200 },
      priority: 2
    }
  ], null, 2)
}

onMounted(() => {
  fetchTemplates()
})
</script>

<template>
  <div class="task-create">
    <h2>创建训练任务</h2>

    <!-- 模式切换 -->
    <div class="mode-switch">
      <button
        :class="['mode-btn', { active: !batchMode }]"
        @click="batchMode = false"
      >
        单个任务
      </button>
      <button
        :class="['mode-btn', { active: batchMode }]"
        @click="batchMode = true"
      >
        批量创建
      </button>
    </div>

    <!-- 成功/错误提示 -->
    <div v-if="success" class="success-message">
      任务创建成功!正在跳转...
    </div>
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 单个任务表单 -->
    <div v-if="!batchMode" class="form-container">
      <!-- 模板选择 -->
      <div v-if="templates.length > 0" class="form-group">
        <label>配置模板</label>
        <div class="template-select">
          <select v-model="selectedTemplate">
            <option value="">手动配置</option>
            <option v-for="tpl in templates" :key="tpl.name" :value="tpl.name">
              {{ tpl.name }}
            </option>
          </select>
          <button @click="applyTemplate" class="btn-secondary">应用</button>
        </div>
      </div>

      <div class="form-group">
        <label>任务名称 *</label>
        <input v-model="taskName" type="text" placeholder="输入任务名称" />
      </div>

      <div class="form-group">
        <label>优先级</label>
        <input v-model.number="taskPriority" type="number" placeholder="0" />
        <span class="hint">数字越大优先级越高</span>
      </div>

      <div class="form-group">
        <label>任务描述</label>
        <textarea v-model="taskDescription" placeholder="输入任务描述(可选)" rows="3"></textarea>
      </div>

      <h3>训练配置</h3>

      <div class="form-row">
        <div class="form-group">
          <label>Hidden Size</label>
          <input v-model.number="hiddenSize" type="number" />
        </div>

        <div class="form-group">
          <label>Learning Rate</label>
          <input v-model.number="learningRate" type="number" step="0.0001" />
        </div>
      </div>

      <div class="form-row">
        <div class="form-group">
          <label>Epochs</label>
          <input v-model.number="epochs" type="number" />
        </div>

        <div class="form-group">
          <label>Layers</label>
          <input v-model.number="layers" type="number" />
        </div>
      </div>

      <div class="form-actions">
        <button
          @click="handleCreateTask"
          :disabled="loading"
          class="btn-primary"
        >
          {{ loading ? '创建中...' : '创建任务' }}
        </button>
      </div>
    </div>

    <!-- 批量创建表单 -->
    <div v-else class="form-container">
      <div class="form-group">
        <label>批量配置 (JSON 格式)</label>
        <textarea
          v-model="batchConfig"
          placeholder="输入 JSON 格式的任务数组"
          rows="15"
          class="batch-textarea"
        ></textarea>
        <button @click="fillExample" class="btn-secondary">填充示例</button>
      </div>

      <div class="form-actions">
        <button
          @click="handleBatchCreate"
          :disabled="loading"
          class="btn-primary"
        >
          {{ loading ? '创建中...' : '批量创建' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.task-create {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.task-create h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
  color: #303133;
}

.task-create h3 {
  margin: 20px 0 15px 0;
  font-size: 16px;
  color: #606266;
}

.mode-switch {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.mode-btn {
  padding: 8px 20px;
  border: 1px solid #dcdfe6;
  background: white;
  color: #606266;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.mode-btn:hover {
  border-color: #409eff;
  color: #409eff;
}

.mode-btn.active {
  background: #409eff;
  color: white;
  border-color: #409eff;
}

.success-message {
  padding: 12px;
  background: #f0f9ff;
  color: #67c23a;
  border-radius: 4px;
  margin-bottom: 20px;
}

.error-message {
  padding: 12px;
  background: #fef0f0;
  color: #f56c6c;
  border-radius: 4px;
  margin-bottom: 20px;
}

.form-container {
  max-width: 600px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #606266;
  font-size: 14px;
  font-weight: 500;
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
  outline: none;
  border-color: #409eff;
}

.hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
}

.template-select {
  display: flex;
  gap: 10px;
}

.template-select select {
  flex: 1;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 15px;
}

.batch-textarea {
  font-family: 'Courier New', monospace;
  font-size: 13px;
  margin-bottom: 10px;
}

.form-actions {
  margin-top: 30px;
  display: flex;
  gap: 10px;
}

.btn-primary,
.btn-secondary {
  padding: 10px 24px;
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

.btn-primary:hover:not(:disabled) {
  background: #66b1ff;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-secondary:hover {
  background: #e4e7ed;
}
</style>
