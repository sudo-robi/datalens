import axios from 'axios'

const API_BASE = '/api'
const INGESTION_BASE = 'http://localhost:8081'

export const api = axios.create({
  baseURL: API_BASE,
})

export const ingestionApi = axios.create({
  baseURL: INGESTION_BASE,
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
    return Promise.reject(error)
  }
)
