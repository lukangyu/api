import axios from 'axios'
import { clearToken, getToken } from '../utils/auth'

const client = axios.create({
  baseURL: '/api',
  timeout: 20000
})

client.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (resp) => resp,
  (err) => {
    if (err?.response?.status === 401) {
      clearToken()
      if (!location.pathname.includes('/login')) {
        location.href = '/login'
      }
    }
    return Promise.reject(err)
  }
)

export default client
