<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElButton, ElDialog, ElMessage, ElTag } from 'element-plus'
import type { WindTurbine } from '@/types'
import { turbineApi } from '@/api'
import { getStatusLabel, getStatusColor } from '@/utils/status'
import TurbineForm from './TurbineForm.vue'

const turbines = ref<WindTurbine[]>([])
const dialogVisible = ref(false)
const editTurbine = ref<WindTurbine | null>(null)

const loadTurbines = async () => {
  try {
    turbines.value = await turbineApi.getAll()
  } catch (error) {
    ElMessage.error('加载风机列表失败')
  }
}

const handleAdd = () => {
  editTurbine.value = null
  dialogVisible.value = true
}

const handleEdit = (turbine: WindTurbine) => {
  editTurbine.value = turbine
  dialogVisible.value = true
}

const handleDelete = async (id: number) => {
  try {
    await turbineApi.delete(id)
    ElMessage.success('删除成功')
    loadTurbines()
  } catch (error) {
    ElMessage.error('删除失败')
  }
}

const handleSubmit = async (data: Omit<WindTurbine, 'id' | 'created_at' | 'updated_at'>) => {
  try {
    if (editTurbine.value) {
      await turbineApi.update(editTurbine.value.id, data)
      ElMessage.success('更新成功')
    } else {
      await turbineApi.create(data)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadTurbines()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

onMounted(loadTurbines)
</script>

<template>
  <div class="turbine-list">
    <div class="header">
      <h2>风机设备列表</h2>
      <ElButton type="primary" @click="handleAdd">添加风机</ElButton>
    </div>
    <ElTable :data="turbines" border stripe>
      <ElTableColumn prop="id" label="ID" width="80" />
      <ElTableColumn prop="name" label="风机名称" />
      <ElTableColumn prop="code" label="设备编号" />
      <ElTableColumn prop="wind_farm" label="所属风场" />
      <ElTableColumn prop="location" label="安装位置" />
      <ElTableColumn prop="status" label="状态">
        <template #default="scope">
          <ElTag :style="{ backgroundColor: getStatusColor(scope.row.status) }">
            {{ getStatusLabel(scope.row.status) }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="capacity" label="装机容量(kW)" />
      <ElTableColumn label="操作" width="200">
        <template #default="scope">
          <ElButton type="primary" link @click="handleEdit(scope.row)">编辑</ElButton>
          <ElButton type="danger" link @click="handleDelete(scope.row.id)">删除</ElButton>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElDialog
      :title="editTurbine ? '编辑风机' : '添加风机'"
      v-model="dialogVisible"
      width="500px"
    >
      <TurbineForm :turbine="editTurbine" @submit="handleSubmit" />
    </ElDialog>
  </div>
</template>

<style scoped>
.turbine-list {
  padding: 20px;
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
}
</style>
