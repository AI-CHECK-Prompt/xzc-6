<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElCard, ElRow, ElCol, ElTable, ElTableColumn, ElTag, ElProgress } from 'element-plus'
import * as echarts from 'echarts'
import type { WindTurbine, TurbineStatus } from '@/types'
import { turbineApi, sensorApi } from '@/api'
import { getStatusLabel, getStatusColor } from '@/utils/status'

const turbines = ref<WindTurbine[]>([])
const statusMap = ref<Map<number, TurbineStatus>>(new Map())
let intervalId: ReturnType<typeof setInterval> | null = null
let chartInstance: echarts.ECharts | null = null

const loadData = async () => {
  try {
    turbines.value = await turbineApi.getAll()
    for (const turbine of turbines.value) {
      try {
        const status = await sensorApi.getStatus(turbine.id)
        statusMap.value.set(turbine.id, status)
      } catch {
        statusMap.value.set(turbine.id, {
          turbine_id: turbine.id,
          rpm: 0,
          power: 0,
          temperature: 0,
          humidity: 0,
          vibration: 0,
          timestamp: '-'
        })
      }
    }
    updateChart()
  } catch (error) {
    console.error('加载数据失败:', error)
  }
}

const initChart = () => {
  const chartDom = document.getElementById('powerChart')
  if (chartDom) {
    chartInstance = echarts.init(chartDom)
    updateChart()
  }
}

const updateChart = () => {
  if (!chartInstance) return

  const names = turbines.value.map(t => t.name)
  const powers = turbines.value.map(t => {
    const status = statusMap.value.get(t.id)
    return status ? status.power : 0
  })

  const option: echarts.EChartsOption = {
    title: {
      text: '风机实时功率'
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: names,
      axisLabel: {
        rotate: 30
      }
    },
    yAxis: {
      type: 'value',
      name: '功率(kW)'
    },
    series: [
      {
        name: '功率',
        type: 'bar',
        data: powers,
        itemStyle: {
          color: '#409eff'
        }
      }
    ]
  }

  chartInstance.setOption(option)
}

const getStatus = (id: number) => statusMap.value.get(id)

onMounted(() => {
  loadData()
  initChart()
  intervalId = setInterval(loadData, 5000)
})

onUnmounted(() => {
  if (intervalId) {
    clearInterval(intervalId)
  }
  if (chartInstance) {
    chartInstance.dispose()
  }
})
</script>

<template>
  <div class="dashboard">
    <h2>实时监控</h2>

    <ElRow :gutter="20">
      <ElCol :span="6">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value">{{ turbines.length }}</div>
            <div class="stat-label">风机总数</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #67c23a">
              {{ turbines.filter(t => t.status === 'running').length }}
            </div>
            <div class="stat-label">运行中</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #e6a23c">
              {{ turbines.filter(t => t.status === 'maintenance').length }}
            </div>
            <div class="stat-label">维护中</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #f56c6c">
              {{ turbines.filter(t => t.status === 'fault').length }}
            </div>
            <div class="stat-label">故障</div>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElRow :gutter="20" style="margin-top: 20px">
      <ElCol :span="16">
        <ElCard>
          <div id="powerChart" style="height: 400px"></div>
        </ElCard>
      </ElCol>
      <ElCol :span="8">
        <ElCard title="设备实时状态">
          <ElTable :data="turbines" stripe size="small">
            <ElTableColumn prop="name" label="风机名称" />
            <ElTableColumn prop="status" label="状态" width="80">
              <template #default="scope">
                <ElTag :style="{ backgroundColor: getStatusColor(scope.row.status) }">
                  {{ getStatusLabel(scope.row.status) }}
                </ElTag>
              </template>
            </ElTableColumn>
            <ElTableColumn label="功率(kW)" width="100">
              <template #default="scope">
                {{ getStatus(scope.row.id)?.power || '-' }}
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElCard style="margin-top: 20px" title="实时数据详情">
      <ElTable :data="turbines" stripe>
        <ElTableColumn prop="name" label="风机名称" />
        <ElTableColumn prop="code" label="设备编号" />
        <ElTableColumn label="转速(RPM)" width="120">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.rpm || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="功率(kW)" width="120">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.power || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="温度(°C)" width="120">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.temperature || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="湿度(%)" width="120">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.humidity || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="振动" width="120">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.vibration || '-' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="更新时间" width="180">
          <template #default="scope">
            {{ getStatus(scope.row.id)?.timestamp || '-' }}
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </div>
</template>

<style scoped>
.dashboard {
  padding: 20px;
}

.dashboard h2 {
  margin: 0 0 20px 0;
  font-size: 20px;
}

.stat-card {
  height: 100%;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
</style>
