import axios from 'axios'
import type { WindTurbine, SensorData, TurbineStatus, TurbineStatistics, SystemStatistics } from '@/types'

const request = axios.create({
  baseURL: '/api',
  timeout: 10000
})

request.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    throw error
  }
)

export const turbineApi = {
  create: (data: Omit<WindTurbine, 'id' | 'created_at' | 'updated_at'>) =>
    request.post<WindTurbine>('/turbines', data),
  getAll: () => request.get<WindTurbine[]>('/turbines'),
  getById: (id: number) => request.get<WindTurbine>(`/turbines/${id}`),
  update: (id: number, data: Partial<WindTurbine>) =>
    request.put<WindTurbine>(`/turbines/${id}`, data),
  delete: (id: number) => request.delete(`/turbines/${id}`)
}

export const sensorApi = {
  collect: (data: Omit<SensorData, 'id' | 'created_at'>) =>
    request.post<SensorData>('/data/collect', data),
  getByTurbineId: (id: number) => request.get<SensorData[]>(`/data/turbine/${id}`),
  getStatus: (id: number) => request.get<TurbineStatus>(`/data/status/${id}`),
  getTrend: (id: number, startTime: string, endTime: string) =>
    request.get<SensorData[]>(`/data/trend/${id}`, { params: { start_time: startTime, end_time: endTime } }),
  getStatistics: (startTime: string, endTime: string) =>
    request.get<TurbineStatistics[]>('/data/statistics', { params: { start_time: startTime, end_time: endTime } }),
  export: (id: number, startTime: string, endTime: string) =>
    request.get(`/data/export/${id}`, { params: { start_time: startTime, end_time: endTime }, responseType: 'blob' })
}

export const statisticsApi = {
  getSystem: () => request.get<SystemStatistics>('/statistics/system')
}
