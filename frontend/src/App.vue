<script setup>
import { ref, onMounted } from 'vue'
// V1 组件
import TaskList from './components/TaskList.vue'
import TaskCreate from './components/TaskCreate.vue'
import TaskDetail from './components/TaskDetail.vue'
import QueueStatus from './components/QueueStatus.vue'
import Statistics from './components/Statistics.vue'
// V2 组件
import V2Dashboard from './components/v2/V2Dashboard.vue'
import { getApiToken, setApiToken } from './api/mlqueue.js'

// API 版本选择
const apiVersion = ref('v2') // 'v1' or 'v2'

// 当前激活的视图
const currentView = ref('tasks')

// API Token
const apiToken = ref('')
const showTokenInput = ref(false)

// 选中的任务ID
const selectedTaskId = ref(null)

// 切换视图
const switchView = (view) => {
  currentView.value = view
  if (view !== 'task-detail') {
    selectedTaskId.value = null
  }
}

// 查看任务详情
const viewTaskDetail = (taskId) => {
  selectedTaskId.value = taskId
  currentView.value = 'task-detail'
}

// 保存 API Token
const saveToken = () => {
  setApiToken(apiToken.value)
  showTokenInput.value = false
}

// 组件挂载时加载 Token
onMounted(() => {
  const token = getApiToken()
  const defaultKey = import.meta.env.VITE_API_KEY

  if (token) {
    apiToken.value = token
  } else if (defaultKey) {
    // 如果环境变量中配置了默认 API Key，自动使用
    apiToken.value = defaultKey
    console.log('使用环境变量中的默认 API Key')
  } else {
    // 没有配置 token 也没有默认 key，显示输入框
    showTokenInput.value = true
  }
})
</script>

<template>
  <div class="app">
    <!-- 顶部导航栏 -->
    <header class="header">
      <div class="header-content">
        <h1 class="logo">MLQueue</h1>

        <!-- V1/V2 版本切换 -->
        <div class="version-switch">
          <button
            :class="['version-btn', { active: apiVersion === 'v1' }]"
            @click="apiVersion = 'v1'"
          >
            V1 (云端调度)
          </button>
          <button
            :class="['version-btn', { active: apiVersion === 'v2' }]"
            @click="apiVersion = 'v2'"
          >
            V2 (Python驱动)
          </button>
        </div>

        <!-- V1 导航菜单 -->
        <nav v-if="apiVersion === 'v1'" class="nav">
          <button
            :class="['nav-btn', { active: currentView === 'tasks' }]"
            @click="switchView('tasks')"
          >
            任务列表
          </button>
          <button
            :class="['nav-btn', { active: currentView === 'create' }]"
            @click="switchView('create')"
          >
            创建任务
          </button>
          <button
            :class="['nav-btn', { active: currentView === 'queue' }]"
            @click="switchView('queue')"
          >
            队列状态
          </button>
          <button
            :class="['nav-btn', { active: currentView === 'statistics' }]"
            @click="switchView('statistics')"
          >
            统计信息
          </button>
        </nav>

        <button class="token-btn" @click="showTokenInput = !showTokenInput">
          {{ apiToken ? '已配置 Token' : '配置 Token' }}
        </button>
      </div>
    </header>

    <!-- Token 配置面板 -->
    <div v-if="showTokenInput" class="token-panel">
      <div class="token-content">
        <h3>配置 API Token</h3>
        <input
          v-model="apiToken"
          type="password"
          placeholder="请输入 API Token"
          class="token-input"
        />
        <div class="token-actions">
          <button @click="saveToken" class="btn-primary">保存</button>
          <button @click="showTokenInput = false" class="btn-secondary">取消</button>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <main class="main-content">
      <!-- V2 架构 -->
      <V2Dashboard v-if="apiVersion === 'v2'" />

      <!-- V1 架构 -->
      <template v-else>
        <TaskList
          v-if="currentView === 'tasks'"
          @view-detail="viewTaskDetail"
        />
        <TaskCreate
          v-else-if="currentView === 'create'"
          @task-created="switchView('tasks')"
        />
        <TaskDetail
          v-else-if="currentView === 'task-detail' && selectedTaskId"
          :task-id="selectedTaskId"
          @back="switchView('tasks')"
        />
        <QueueStatus v-else-if="currentView === 'queue'" />
        <Statistics v-else-if="currentView === 'statistics'" />
      </template>
    </main>
  </div>
</template>

<style scoped>
.app {
  min-height: 100vh;
  background: #f5f7fa;
}

.header {
  background: white;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
  display: flex;
  align-items: center;
  height: 60px;
}

.logo {
  font-size: 24px;
  font-weight: 600;
  color: #409eff;
  margin: 0;
  margin-right: 20px;
}

.version-switch {
  display: flex;
  gap: 5px;
  margin-right: 20px;
  padding: 4px;
  background: #f5f7fa;
  border-radius: 6px;
}

.version-btn {
  padding: 6px 12px;
  border: none;
  background: transparent;
  color: #606266;
  cursor: pointer;
  border-radius: 4px;
  font-size: 13px;
  transition: all 0.3s;
  white-space: nowrap;
}

.version-btn:hover {
  color: #409eff;
}

.version-btn.active {
  background: white;
  color: #409eff;
  font-weight: 600;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.nav {
  display: flex;
  gap: 10px;
  flex: 1;
}

.nav-btn {
  padding: 8px 16px;
  border: none;
  background: transparent;
  color: #606266;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.3s;
}

.nav-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

.nav-btn.active {
  background: #409eff;
  color: white;
}

.token-btn {
  padding: 8px 16px;
  border: 1px solid #dcdfe6;
  background: white;
  color: #606266;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.3s;
}

.token-btn:hover {
  border-color: #409eff;
  color: #409eff;
}

.token-panel {
  background: rgba(0, 0, 0, 0.5);
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.token-content {
  background: white;
  padding: 30px;
  border-radius: 8px;
  width: 400px;
}

.token-content h3 {
  margin: 0 0 20px 0;
  font-size: 18px;
  color: #303133;
}

.token-input {
  width: 100%;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  margin-bottom: 20px;
  box-sizing: border-box;
}

.token-input:focus {
  outline: none;
  border-color: #409eff;
}

.token-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
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
  background: #f5f7fa;
  color: #606266;
}

.btn-secondary:hover {
  background: #e4e7ed;
}

.main-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
</style>
