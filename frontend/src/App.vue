<script setup lang="ts">
import { ref } from 'vue'
import { ElContainer, ElAside, ElMenu, ElMenuItem, ElMain } from 'element-plus'
import TurbineList from './components/TurbineList.vue'
import MonitorDashboard from './components/MonitorDashboard.vue'
import TrendAnalysis from './components/TrendAnalysis.vue'
import StatisticsAnalysis from './components/StatisticsAnalysis.vue'

const activeMenu = ref('monitor')

const menuItems = [
  { key: 'monitor', label: '实时监控' },
  { key: 'turbine', label: '设备管理' },
  { key: 'trend', label: '趋势分析' },
  { key: 'statistics', label: '统计分析' }
]
</script>

<template>
  <ElContainer style="height: 100vh">
    <ElAside width="200px" style="background-color: #304156">
      <div style="color: white; font-size: 20px; padding: 20px; text-align: center;">
        风电运维监控
      </div>
      <ElMenu
        :default-active="activeMenu"
        class="el-menu-vertical-demo"
        style="background-color: #304156; color: white;"
        text-color="white"
        active-text-color="#409eff"
        @select="activeMenu = $event"
      >
        <ElMenuItem v-for="item in menuItems" :key="item.key">
          {{ item.label }}
        </ElMenuItem>
      </ElMenu>
    </ElAside>
    <ElMain>
      <MonitorDashboard v-if="activeMenu === 'monitor'" />
      <TurbineList v-else-if="activeMenu === 'turbine'" />
      <TrendAnalysis v-else-if="activeMenu === 'trend'" />
      <StatisticsAnalysis v-else-if="activeMenu === 'statistics'" />
    </ElMain>
  </ElContainer>
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

html, body, #app {
  height: 100%;
}

.el-menu-vertical-demo {
  border-right: none;
}
</style>
