import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:3000',
  timeout: 5000,
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    console.error('API error:', err.message)
    return Promise.reject(err)
  }
)

export default api