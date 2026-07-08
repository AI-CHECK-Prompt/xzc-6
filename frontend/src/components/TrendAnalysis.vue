<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElCard, ElRow, ElCol, ElDatePicker, ElSelect, ElOption, ElButton, ElMessage, ElDownload } from 'element-plus'
import * as echarts from 'echarts'
import type { WindTurbine, SensorData } from '@/types'
import { turbineApi, sensorApi } from '@/api'

const turbines = ref<WindTurbine[]>([])
const selectedTurbine = ref<number | null>(null)
const startTime = ref('')
const endTime = ref('')
const trendData = ref<SensorData[]>([])
const activeMetric = ref('power')

let chartInstance: echarts.ECharts | null = null

const metrics = [
  { key: 'power', label: '功率(kW)' },
  { key: 'rpm', label: '转速(RPM)' },
  { key: 'temperature', label: '温度(°C)' },
  { key: 'humidity', label: '湿度(%)' },
  { key: 'vibration', label: '振动' }
]

const loadTurbines = async () => {
  try {
    turbines.value = await turbineApi.getAll()
    if (turbines.value.length > 0) {
      selectedTurbine.value = turbines.value[0].id
    }
  } catch (error) {
    ElMessage.error('加载风机列表失败')
  }
}

const loadTrendData = async () => {
  if (!selectedTurbine.value || !startTime.value || !endTime.value) {
    ElMessage.warning('请选择风机和时间范围')
    return
  }

  try {
    trendData.value = await sensorApi.getTrend(selectedTurbine.value, startTime.value, endTime.value)
    updateChart()
  } catch (error) {
    ElMessage.error('加载趋势数据失败')
  }
}

const exportData = async () => {
  if (!selectedTurbine.value || !startTime.value || !endTime.value) {
    ElMessage.warning('请选择风机和时间范围')
    return
  }

  try {
    const blob = await sensorApi.export(selectedTurbine.value, startTime.value, endTime.value)
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `turbine_${selectedTurbine.value}_data.xlsx`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

const initChart = () => {
  const chartDom = document.getElementById('trendChart')
  if (chartDom) {
    chartInstance = echarts.init(chartDom)
  }
}

const updateChart = () => {
  if (!chartInstance || trendData.value.length === 0) return

  const times = trendData.value.map(d => {
    const date = new Date(d.created_at)
    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
  })
  const values = trendData.value.map(d => d[activeMetric.value as keyof SensorData])

  const option: echarts.EChartsOption = {
    title: {
      text: `${metrics.find(m => m.key === activeMetric.value)?.label}趋势`
    },
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: times,
      axisLabel: {
        rotate: 45
      }
    },
    yAxis: {
      type: 'value',
      name: metrics.find(m => m.key === activeMetric.value)?.label
    },
    series: [
      {
        name: metrics.find(m => m.key === activeMetric.value)?.label,
        type: 'line',
        smooth: true,
        data: values,
        lineStyle: {
          width: 2
        },
        areaStyle: {
          opacity: 0.3
        }
      }
    ]
  }

  chartInstance.setOption(option)
}

watch(activeMetric, () => {
  updateChart()
})

onMounted(() => {
  loadTurbines()
  initChart()
})
</script>

<template>
  <div class="trend-analysis">
    <h2>趋势分析</h2>

    <ElCard style="margin-bottom: 20px">
      <ElRow :gutter="20">
        <ElCol :span="6">
          <div class="form-item">
            <label>选择风机</label>
            <ElSelect v-model="selectedTurbine" placeholder="请选择风机">
              <ElOption v-for="turbine in turbines" :key="turbine.id" :label="turbine.name" :value="turbine.id" />
            </ElSelect>
          </div>
        </ElCol>
        <ElCol :span="6">
          <div class="form-item">
            <label>开始时间</label>
            <ElDatePicker v-model="startTime" type="datetime" placeholder="选择开始时间" />
          </div>
        </ElCol>
        <ElCol :span="6">
          <div class="form-item">
            <label>结束时间</label>
            <ElDatePicker v-model="endTime" type="datetime" placeholder="选择结束时间" />
          </div>
        </ElCol>
        <ElCol :span="6">
          <div class="form-item">
            <label>操作</label>
            <div style="display: flex; gap: 10px">
              <ElButton type="primary" @click="loadTrendData">查询</ElButton>
              <ElButton type="success" @click="exportData">
                <ElDownload /> 导出
              </ElButton>
            </div>
          </div>
        </ElCol>
      </ElRow>
    </ElCard>

    <ElCard style="margin-bottom: 20px">
      <ElRow :gutter="20">
        <ElCol :span="12">
          <div class="form-item">
            <label>选择指标</label>
            <ElSelect v-model="activeMetric" placeholder="请选择指标">
              <ElOption v-for="metric in metrics" :key="metric.key" :label="metric.label" :value="metric.key" />
            </ElSelect>
          </div>
        </ElCol>
      </ElRow>
    </ElCard>

    <ElCard>
      <div id="trendChart" style="height: 500px"></div>
    </ElCard>
  </div>
</template>

<style scoped>
.trend-analysis {
  padding: 20px;
}

.trend-analysis h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
}

.form-item {
  display: flex;
  flex-direction: column;
}

.form-item label {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}
</style>
