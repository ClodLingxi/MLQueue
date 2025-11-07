<script setup>
import { ref, onMounted } from 'vue'
import {
  listGroups,
  createGroup,
  updateGroup,
  deleteGroup,
  getGroup
} from '@/api/mlqueue-v2.js'

const emit = defineEmits(['select-group'])

// 数据
const groups = ref([])
const loading = ref(false)
const error = ref(null)

// 对话框状态
const showCreateDialog = ref(false)
const showEditDialog = ref(false)

// 表单数据
const createForm = ref({
  name: '',
  description: ''
})

const editForm = ref({
  group_id: '',
  name: '',
  description: ''
})

// 加载组列表
const fetchGroups = async () => {
  loading.value = true
  error.value = null
  try {
    const response = await listGroups()
    groups.value = response.groups || []
  } catch (err) {
    error.value = err.message || '获取组列表失败'
    console.error('获取组列表失败:', err)
  } finally {
    loading.value = false
  }
}

// 创建组
const handleCreate = async () => {
  if (!createForm.value.name) {
    alert('请输入组名称')
    return
  }

  try {
    await createGroup(createForm.value.name, createForm.value.description)
    showCreateDialog.value = false
    createForm.value = { name: '', description: '' }
    await fetchGroups()
  } catch (err) {
    alert('创建组失败: ' + err.message)
  }
}

// 打开编辑对话框
const openEditDialog = async (group) => {
  try {
    const response = await getGroup(group.group_id)
    editForm.value = {
      group_id: response.group.group_id,
      name: response.group.name,
      description: response.group.description || ''
    }
    showEditDialog.value = true
  } catch (err) {
    alert('获取组详情失败: ' + err.response?.data?.error || err.message)
  }
}

// 更新组
const handleUpdate = async () => {
  if (!editForm.value.name) {
    alert('请输入组名称')
    return
  }

  try {
    await updateGroup(editForm.value.group_id, {
      name: editForm.value.name,
      description: editForm.value.description
    })
    showEditDialog.value = false
    await fetchGroups()
  } catch (err) {
    alert('更新组失败: ' + err.response?.data?.error || err.message)
  }
}

// 删除组
const handleDelete = async (groupId) => {
  if (!confirm('确定要删除这个组吗？删除后无法恢复！')) return

  try {
    await deleteGroup(groupId)
    await fetchGroups()
  } catch (err) {
    alert('删除组失败: ' + err.response?.data?.error || err.message)
  }
}

// 选择组
const handleSelectGroup = (group) => {
  emit('select-group', group)
}

// 格式化时间
const formatDate = (dateString) => {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

onMounted(() => {
  fetchGroups()
})
</script>

<template>
  <div class="group-management">
    <div class="header">
      <h2>项目组管理</h2>
      <div class="actions">
        <button @click="fetchGroups" class="btn-refresh">刷新</button>
        <button @click="showCreateDialog = true" class="btn-create">创建组</button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading">加载中...</div>

    <!-- 组列表 -->
    <div v-else-if="groups.length > 0" class="groups-grid">
      <div
        v-for="group in groups"
        :key="group.group_id"
        class="group-card"
        @click="handleSelectGroup(group)"
      >
        <div class="card-header">
          <h3>{{ group.name }}</h3>
          <div class="card-actions" @click.stop>
            <button @click="openEditDialog(group)" class="btn-icon" title="编辑">
              ✏️
            </button>
            <button @click="handleDelete(group.group_id)" class="btn-icon" title="删除">
              🗑️
            </button>
          </div>
        </div>
        <p class="card-description">{{ group.description || '暂无描述' }}</p>
        <div class="card-footer">
          <span class="card-meta">创建于 {{ formatDate(group.created_at) }}</span>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <p>暂无项目组</p>
      <button @click="showCreateDialog = true" class="btn-primary">创建第一个组</button>
    </div>

    <!-- 创建对话框 -->
    <div v-if="showCreateDialog" class="dialog-overlay" @click.self="showCreateDialog = false">
      <div class="dialog">
        <h3>创建项目组</h3>
        <div class="form-group">
          <label>组名称 *</label>
          <input
            v-model="createForm.name"
            type="text"
            placeholder="例如：MNIST训练项目"
            @keyup.enter="handleCreate"
          />
        </div>
        <div class="form-group">
          <label>描述</label>
          <textarea
            v-model="createForm.description"
            placeholder="项目描述（可选）"
            rows="3"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="showCreateDialog = false" class="btn-secondary">取消</button>
          <button @click="handleCreate" class="btn-primary">创建</button>
        </div>
      </div>
    </div>

    <!-- 编辑对话框 -->
    <div v-if="showEditDialog" class="dialog-overlay" @click.self="showEditDialog = false">
      <div class="dialog">
        <h3>编辑项目组</h3>
        <div class="form-group">
          <label>组名称 *</label>
          <input
            v-model="editForm.name"
            type="text"
            @keyup.enter="handleUpdate"
          />
        </div>
        <div class="form-group">
          <label>描述</label>
          <textarea
            v-model="editForm.description"
            rows="3"
          ></textarea>
        </div>
        <div class="dialog-actions">
          <button @click="showEditDialog = false" class="btn-secondary">取消</button>
          <button @click="handleUpdate" class="btn-primary">更新</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.group-management {
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
.btn-create {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s;
}

.btn-refresh {
  background: #f5f7fa;
  color: #606266;
  border: 1px solid #dcdfe6;
}

.btn-refresh:hover {
  background: #e4e7ed;
}

.btn-create {
  background: #409eff;
  color: white;
}

.btn-create:hover {
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

.groups-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.group-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s;
}

.group-card:hover {
  border-color: #409eff;
  box-shadow: 0 4px 8px rgba(64, 158, 255, 0.1);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.card-header h3 {
  margin: 0;
  font-size: 18px;
  color: #303133;
  flex: 1;
}

.card-actions {
  display: flex;
  gap: 8px;
}

.btn-icon {
  background: none;
  border: none;
  font-size: 16px;
  cursor: pointer;
  padding: 4px;
  opacity: 0.6;
  transition: opacity 0.3s;
}

.btn-icon:hover {
  opacity: 1;
}

.card-description {
  color: #606266;
  font-size: 14px;
  margin: 0 0 12px 0;
  line-height: 1.6;
}

.card-footer {
  border-top: 1px solid #f5f7fa;
  padding-top: 12px;
}

.card-meta {
  font-size: 12px;
  color: #909399;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
}

.empty-state p {
  color: #909399;
  font-size: 16px;
  margin-bottom: 20px;
}

/* 对话框样式 */
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
  max-height: 90vh;
  overflow-y: auto;
}

.dialog h3 {
  margin: 0 0 20px 0;
  font-size: 18px;
  color: #303133;
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
.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
  font-family: inherit;
}

.form-group input:focus,
.form-group textarea:focus {
  outline: none;
  border-color: #409eff;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 24px;
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
  border: 1px solid #dcdfe6;
}

.btn-secondary:hover {
  background: #e4e7ed;
}
</style>
