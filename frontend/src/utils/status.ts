import type { WindTurbine } from '@/types'

export const statusMap: Record<WindTurbine['status'], { label: string; color: string }> = {
  running: { label: '运行中', color: '#67c23a' },
  stopped: { label: '已停机', color: '#909399' },
  maintenance: { label: '维护中', color: '#e6a23c' },
  fault: { label: '故障', color: '#f56c6c' }
}

export const getStatusLabel = (status: WindTurbine['status']) => statusMap[status].label
export const getStatusColor = (status: WindTurbine['status']) => statusMap[status].color
