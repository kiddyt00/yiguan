const BASE = 'https://49.235.108.61/api'

function request(path, options = {}) {
  const token = uni.getStorageSync('token')
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`

  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE + path,
      method: options.method || 'GET',
      data: options.data,
      header: headers,
      success(res) {
        if (res.statusCode === 401) {
          uni.removeStorageSync('token')
          uni.reLaunch({ url: '/pages/login/login' })
          return
        }
        if (res.statusCode >= 400) {
          reject(new Error(res.data?.error || '请求失败'))
          return
        }
        resolve(res.data)
      },
      fail(err) {
        reject(new Error('网络错误: ' + err.errMsg))
      }
    })
  })
}

export const api = {
  // 认证
  register: (phone, password) => request('/auth/register', { method: 'POST', data: { phone, password, nickname: '易友' } }),
  login: (phone, password) => request('/auth/login', { method: 'POST', data: { phone, password } }),

  // 用户
  profile: () => request('/user'),
  updateProfile: (data) => request('/user', { method: 'PUT', data }),

  // 起卦
  divine: (question) => request('/divine', { method: 'POST', data: { question } }),
  divineStream: (question) => request('/divine/stream', { method: 'POST', data: { question } }),

  // 历史
  history: (limit = 20, offset = 0) => request(`/history?limit=${limit}&offset=${offset}`),

  // 广告
  activeAds: () => request('/ads/active'),
  watchAd: (id) => request(`/ads/${id}/watch`, { method: 'POST' }),
  completeAd: (id, duration) => request(`/ads/${id}/complete`, { method: 'POST', data: { duration } }),
}
