<script setup>
import { ref } from 'vue'
import GroupManagement from './GroupManagement.vue'
import UnitManagement from './UnitManagement.vue'
import QueueManagement from './QueueManagement.vue'

const currentView = ref('groups') // groups | units | queues
const selectedGroup = ref(null)
const selectedUnit = ref(null)

const handleSelectGroup = (group) => {
  selectedGroup.value = group
  currentView.value = 'units'
}

const handleSelectUnit = (unit) => {
  selectedUnit.value = unit
  currentView.value = 'queues'
}

const handleBackToGroups = () => {
  selectedGroup.value = null
  selectedUnit.value = null
  currentView.value = 'groups'
}

const handleBackToUnits = () => {
  selectedUnit.value = null
  currentView.value = 'units'
}
</script>

<template>
  <div class="v2-dashboard">
    <!-- 面包屑导航 -->
    <div class="breadcrumb">
      <span
        :class="['breadcrumb-item', { active: currentView === 'groups' }]"
        @click="handleBackToGroups"
      >
        项目组
      </span>
      <span v-if="selectedGroup" class="breadcrumb-separator">/</span>
      <span
        v-if="selectedGroup"
        :class="['breadcrumb-item', { active: currentView === 'units' }]"
        @click="handleBackToUnits"
      >
        {{ selectedGroup.name }}
      </span>
      <span v-if="selectedUnit" class="breadcrumb-separator">/</span>
      <span
        v-if="selectedUnit"
        :class="['breadcrumb-item', { active: currentView === 'queues' }]"
      >
        {{ selectedUnit.name }}
      </span>
    </div>

    <!-- 内容区域 -->
    <div class="content">
      <GroupManagement
        v-if="currentView === 'groups'"
        @select-group="handleSelectGroup"
      />
      <UnitManagement
        v-else-if="currentView === 'units' && selectedGroup"
        :group-id="selectedGroup.group_id"
        @select-unit="handleSelectUnit"
        @back="handleBackToGroups"
      />
      <QueueManagement
        v-else-if="currentView === 'queues' && selectedUnit"
        :unit-id="selectedUnit.unit_id"
        @back="handleBackToUnits"
      />
    </div>
  </div>
</template>

<style scoped>
.v2-dashboard {
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 20px;
  padding: 15px 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.breadcrumb-item {
  font-size: 14px;
  color: #606266;
  cursor: pointer;
  transition: color 0.3s;
}

.breadcrumb-item:hover {
  color: #409eff;
}

.breadcrumb-item.active {
  color: #303133;
  font-weight: 600;
  cursor: default;
}

.breadcrumb-separator {
  margin: 0 10px;
  color: #c0c4cc;
}

.content {
  margin-top: 20px;
}
</style>
