function getAuthHeader() {
  const token = localStorage.getItem('admin_token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

async function api(path, options = {}) {
  const headers = { 'Content-Type': 'application/json', ...getAuthHeader() }
  const res = await fetch(`/api${path}`, { ...options, headers })
  if (!res.ok) {
    if (res.status === 401) {
      const { logout } = await import('../stores/auth')
      logout()
      window.location.replace('/admin/login')
      throw new Error('登录已过期')
    }
    const err = await res.json().catch(() => ({ error: '请求失败' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export const adminApi = {
  login: (data) => fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  }).then(r => r.json()),

  dashboard: () => api('/admin/dashboard'),
  dashboardData: () => api('/admin/dashboard'),

  users: (params) => api(`/admin/users?limit=${params.limit || 50}&offset=${params.offset || 0}${params.keyword ? '&keyword=' + encodeURIComponent(params.keyword) : ''}`),
  toggleUser: (id) => api(`/admin/users/${id}/toggle`, { method: 'POST' }),
  adjustQuota: (id, delta) => api(`/admin/users/${id}/quota`, { method: 'POST', body: JSON.stringify({ delta }) }),
  userHistory: (id, limit = 20, offset = 0) => api(`/admin/users/${id}/history?limit=${limit}&offset=${offset}`),

  hexagrams: (params) => {
    let url = `/admin/hexagrams?limit=${params.limit || 20}&offset=${params.offset || 0}`
    if (params.user_id) url += '&user_id=' + params.user_id
    if (params.keyword) url += '&keyword=' + encodeURIComponent(params.keyword)
    if (params.date_from) url += '&date_from=' + params.date_from
    if (params.date_to) url += '&date_to=' + params.date_to
    return api(url)
  },
  hexagramDetail: (id) => api(`/admin/hexagrams/${id}`),
  deleteHexagram: (id) => api(`/admin/hexagrams/${id}`, { method: 'DELETE' }),

  models: () => api('/admin/models'),
  createModel: (data) => api('/admin/models', { method: 'POST', body: JSON.stringify(data) }),
  updateModel: (id, data) => api(`/admin/models/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteModel: (id) => api(`/admin/models/${id}`, { method: 'DELETE' }),
  setDefaultModel: (id) => api(`/admin/models/${id}/set-default`, { method: 'POST' }),
  toggleModel: (id, enabled) => api(`/admin/models/${id}/toggle?enabled=${enabled}`, { method: 'POST' }),
  fetchModels: (endpoint, api_key) => api('/admin/models/fetch', { method: 'POST', body: JSON.stringify({ endpoint, api_key }) }),
  testConnection: (endpoint, api_key, model) => api('/admin/models/test', { method: 'POST', body: JSON.stringify({ endpoint, api_key, model }) }),

  analytics: () => api('/admin/analytics'),

  ads: () => api('/admin/ads'),
  createAd: (data) => api('/admin/ads', { method: 'POST', body: JSON.stringify(data) }),
  updateAd: (id, data) => api(`/admin/ads/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteAd: (id) => api(`/admin/ads/${id}`, { method: 'DELETE' }),
  toggleAd: (id, enabled) => api(`/admin/ads/${id}/toggle?enabled=${enabled}`, { method: 'POST' }),
  adStats: () => api('/admin/ads/stats'),
}
