export interface WindTurbine {
  id: number
  name: string
  code: string
  wind_farm: string
  location: string
  status: 'running' | 'stopped' | 'maintenance' | 'fault'
  install_date: string
  capacity: number
  created_at: string
  updated_at: string
}

export interface SensorData {
  id: number
  turbine_id: number
  timestamp: string
  rpm: number
  power: number
  temperature: number
  humidity: number
  vibration: number
  created_at: string
}

export interface TurbineStatus {
  turbine_id: number
  rpm: number
  power: number
  temperature: number
  humidity: number
  vibration: number
  timestamp: string
}

export interface TurbineStatistics {
  turbine_id: number
  count: number
  avg_power: number
  max_power: number
  min_power: number
  avg_temperature: number
  avg_vibration: number
}

export interface SystemStatistics {
  total_turbines: number
  running_turbines: number
  fault_turbines: number
  maintenance_count: number
  avg_power: number
  total_power: number
}

export interface ApiResponse<T = any> {
  data: T
  code: number
  message: string
}
