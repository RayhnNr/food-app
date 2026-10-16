import axios from 'axios'

const api = axios.create({
//   baseURL: 'http://localhost:3000',
  baseURL: 'http://localhost:8080',
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