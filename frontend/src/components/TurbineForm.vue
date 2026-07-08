<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElForm, ElFormItem, ElInput, ElSelect, ElOption, ElDatePicker, ElButton } from 'element-plus'
import type { WindTurbine } from '@/types'

const props = defineProps<{
  turbine: WindTurbine | null
}>()

const emit = defineEmits<{
  (e: 'submit', data: Omit<WindTurbine, 'id' | 'created_at' | 'updated_at'>): void
}>()

const form = ref({
  name: '',
  code: '',
  wind_farm: '',
  location: '',
  status: 'running' as WindTurbine['status'],
  install_date: '',
  capacity: 0
})

watch(() => props.turbine, (newVal) => {
  if (newVal) {
    form.value = {
      name: newVal.name,
      code: newVal.code,
      wind_farm: newVal.wind_farm,
      location: newVal.location,
      status: newVal.status,
      install_date: newVal.install_date,
      capacity: newVal.capacity
    }
  } else {
    form.value = {
      name: '',
      code: '',
      wind_farm: '',
      location: '',
      status: 'running',
      install_date: '',
      capacity: 0
    }
  }
}, { immediate: true })

const handleSubmit = () => {
  emit('submit', { ...form.value })
}
</script>

<template>
  <ElForm :model="form" label-width="100px">
    <ElFormItem label="风机名称" prop="name">
      <ElInput v-model="form.name" placeholder="请输入风机名称" />
    </ElFormItem>
    <ElFormItem label="设备编号" prop="code">
      <ElInput v-model="form.code" placeholder="请输入设备编号" />
    </ElFormItem>
    <ElFormItem label="所属风场" prop="wind_farm">
      <ElInput v-model="form.wind_farm" placeholder="请输入所属风场" />
    </ElFormItem>
    <ElFormItem label="安装位置" prop="location">
      <ElInput v-model="form.location" placeholder="请输入安装位置" />
    </ElFormItem>
    <ElFormItem label="状态" prop="status">
      <ElSelect v-model="form.status">
        <ElOption label="运行中" value="running" />
        <ElOption label="已停机" value="stopped" />
        <ElOption label="维护中" value="maintenance" />
        <ElOption label="故障" value="fault" />
      </ElSelect>
    </ElFormItem>
    <ElFormItem label="安装日期" prop="install_date">
      <ElDatePicker v-model="form.install_date" type="date" placeholder="请选择安装日期" />
    </ElFormItem>
    <ElFormItem label="装机容量(kW)" prop="capacity">
      <ElInput v-model.number="form.capacity" placeholder="请输入装机容量" />
    </ElFormItem>
    <ElFormItem>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
      <ElButton @click="$emit('update:visible', false)">取消</ElButton>
    </ElFormItem>
  </ElForm>
</template>
