<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElCard, ElRow, ElCol, ElDatePicker, ElButton, ElMessage, ElTable, ElTableColumn } from 'element-plus'
import * as echarts from 'echarts'
import type { SystemStatistics, TurbineStatistics, WindTurbine } from '@/types'
import { statisticsApi, sensorApi, turbineApi } from '@/api'

const systemStats = ref<SystemStatistics | null>(null)
const turbineStats = ref<TurbineStatistics[]>([])
const turbines = ref<WindTurbine[]>([])
const startTime = ref('')
const endTime = ref('')

let chartInstance: echarts.ECharts | null = null

const loadSystemStats = async () => {
  try {
    systemStats.value = await statisticsApi.getSystem()
  } catch (error) {
    ElMessage.error('加载系统统计失败')
  }
}

const loadTurbineStats = async () => {
  if (!startTime.value || !endTime.value) {
    ElMessage.warning('请选择时间范围')
    return
  }

  try {
    turbineStats.value = await sensorApi.getStatistics(startTime.value, endTime.value)
    updateChart()
  } catch (error) {
    ElMessage.error('加载风机统计失败')
  }
}

const loadTurbines = async () => {
  try {
    turbines.value = await turbineApi.getAll()
  } catch (error) {
    ElMessage.error('加载风机列表失败')
  }
}

const initChart = () => {
  const chartDom = document.getElementById('statsChart')
  if (chartDom) {
    chartInstance = echarts.init(chartDom)
  }
}

const updateChart = () => {
  if (!chartInstance || turbineStats.value.length === 0) return

  const turbineNames = turbineStats.value.map(s => {
    const turbine = turbines.value.find(t => t.id === s.turbine_id)
    return turbine ? turbine.name : `风机${s.turbine_id}`
  })
  const avgPowers = turbineStats.value.map(s => s.avg_power)
  const maxPowers = turbineStats.value.map(s => s.max_power)

  const option: echarts.EChartsOption = {
    title: {
      text: '风机功率对比'
    },
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['平均功率', '最大功率']
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: turbineNames,
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
        name: '平均功率',
        type: 'bar',
        data: avgPowers,
        itemStyle: {
          color: '#409eff'
        }
      },
      {
        name: '最大功率',
        type: 'bar',
        data: maxPowers,
        itemStyle: {
          color: '#67c23a'
        }
      }
    ]
  }

  chartInstance.setOption(option)
}

const getTurbineName = (id: number) => {
  const turbine = turbines.value.find(t => t.id === id)
  return turbine ? turbine.name : `风机${id}`
}

onMounted(() => {
  loadSystemStats()
  loadTurbines()
  initChart()
})
</script>

<template>
  <div class="statistics-analysis">
    <h2>统计分析</h2>

    <ElRow :gutter="20" style="margin-bottom: 20px">
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value">{{ systemStats?.total_turbines || 0 }}</div>
            <div class="stat-label">风机总数</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #67c23a">{{ systemStats?.running_turbines || 0 }}</div>
            <div class="stat-label">运行中</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #e6a23c">{{ systemStats?.maintenance_count || 0 }}</div>
            <div class="stat-label">维护中</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value" style="color: #f56c6c">{{ systemStats?.fault_turbines || 0 }}</div>
            <div class="stat-label">故障</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value">{{ (systemStats?.avg_power || 0).toFixed(2) }}</div>
            <div class="stat-label">平均功率(kW)</div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :span="4">
        <ElCard class="stat-card">
          <div class="stat-item">
            <div class="stat-value">{{ (systemStats?.total_power || 0).toFixed(2) }}</div>
            <div class="stat-label">总功率(kW)</div>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElCard style="margin-bottom: 20px">
      <ElRow :gutter="20">
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
            <ElButton type="primary" @click="loadTurbineStats">查询</ElButton>
          </div>
        </ElCol>
      </ElRow>
    </ElCard>

    <ElRow :gutter="20">
      <ElCol :span="14">
        <ElCard>
          <div id="statsChart" style="height: 400px"></div>
        </ElCard>
      </ElCol>
      <ElCol :span="10">
        <ElCard title="风机统计详情">
          <ElTable :data="turbineStats" stripe size="small">
            <ElTableColumn label="风机名称" width="120">
              <template #default="scope">
                {{ getTurbineName(scope.row.turbine_id) }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="avg_power" label="平均功率" width="100">
              <template #default="scope">
                {{ scope.row.avg_power.toFixed(2) }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="max_power" label="最大功率" width="100">
              <template #default="scope">
                {{ scope.row.max_power.toFixed(2) }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="min_power" label="最小功率" width="100">
              <template #default="scope">
                {{ scope.row.min_power.toFixed(2) }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="avg_temperature" label="平均温度" width="100">
              <template #default="scope">
                {{ scope.row.avg_temperature.toFixed(2) }}
              </template>
            </ElTableColumn>
            <ElTableColumn prop="avg_vibration" label="平均振动" width="100">
              <template #default="scope">
                {{ scope.row.avg_vibration.toFixed(2) }}
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<style scoped>
.statistics-analysis {
  padding: 20px;
}

.statistics-analysis h2 {
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
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
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
