import axios from 'axios'

const API_BASE = import.meta.env.VITE_API_URL || '/api'
const INGESTION_BASE = import.meta.env.VITE_INGESTION_URL || 'http://localhost:8081'

export const api = axios.create({
  baseURL: API_BASE,
  timeout: 10000,
})

export const ingestionApi = axios.create({
  baseURL: INGESTION_BASE,
  timeout: 10000,
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    if (error.code === 'ERR_NETWORK' || error.response?.status === 404) {
      return Promise.reject(new Error('API unavailable. The backend server is not running.'))
    }
    return Promise.reject(error)
  }
)
