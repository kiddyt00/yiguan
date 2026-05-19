const config = require('./config.js').default || require('./config.js')
const BASE = config.API_BASE

function request(path, options = {}) {
  const token = wx.getStorageSync('token')
  const header = { 'Content-Type': 'application/json' }
  if (token) header['Authorization'] = 'Bearer ' + token

  return new Promise((resolve, reject) => {
    wx.request({
      url: BASE + path,
      method: options.method || 'GET',
      data: options.data,
      header,
      timeout: options.timeout || 30000,
      success(res) {
        if (res.statusCode === 401) {
          wx.removeStorageSync('token')
          wx.removeStorageSync('user')
          wx.showToast({ title: '登录已过期', icon: 'none' })
          wx.reLaunch({ url: '/pages/login/login' })
          return
        }
        if (res.statusCode >= 400) {
          reject(new Error(res.data?.error || '请求失败(' + res.statusCode + ')'))
          return
        }
        resolve(res.data)
      },
      fail(err) {
        reject(new Error('网络错误: ' + (err.errMsg || '连接失败')))
      }
    })
  })
}

module.exports = {
  register: (phone, password) => request('/auth/register', { method: 'POST', data: { phone, password, nickname: '易友' } }),
  login: (phone, password) => request('/auth/login', { method: 'POST', data: { phone, password } }),
  profile: () => request('/user'),
  updateProfile: (data) => request('/user', { method: 'PUT', data }),
  divine: (question) => request('/divine', { method: 'POST', data: { question } }),
  history: (limit = 20, offset = 0) => request('/history?limit=' + limit + '&offset=' + offset),
  activeAds: () => request('/ads/active'),
  watchAd: (id) => request('/ads/' + id + '/watch', { method: 'POST' }),
  completeAd: (id, duration) => request('/ads/' + id + '/complete', { method: 'POST', data: { duration } }),
}
