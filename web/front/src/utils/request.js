// 统一 API 请求封装
// - 自动注入 Bearer token
// - 401 时自动跳转登录页
// - 统一 JSON 解析

const BASE_URL = ''

async function request(url, options = {}) {
  const token = localStorage.getItem('token')

  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  }

  if (token) {
    headers['Authorization'] = 'Bearer ' + token
  }

  const res = await fetch(BASE_URL + url, {
    ...options,
    headers,
  })

  // token 过期或无效，清除登录态并跳转
  if (res.status === 401) {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    window.location.href = '/login'
    throw new Error('未登录')
  }

  return res
}

// GET 请求
export function apiGet(url, params) {
  let fullUrl = url
  if (params) {
    const qs = new URLSearchParams(params).toString()
    fullUrl += (url.includes('?') ? '&' : '?') + qs
  }
  return request(fullUrl)
}

// POST 请求（headers 可选，用于传 Accept-Language 等）
export function apiPost(url, data, headers) {
  return request(url, {
    method: 'POST',
    body: JSON.stringify(data),
    headers,
  })
}

// PUT 请求（headers 可选）
export function apiPut(url, data, headers) {
  return request(url, {
    method: 'PUT',
    body: JSON.stringify(data),
    headers,
  })
}

// DELETE 请求
export function apiDelete(url) {
  return request(url, { method: 'DELETE' })
}

// 快捷方法：直接返回 JSON
export function apiGetJSON(url, params) {
  return apiGet(url, params).then(r => r.json())
}

export function apiPostJSON(url, data, headers) {
  return apiPost(url, data, headers).then(r => r.json())
}

export function apiPutJSON(url, data, headers) {
  return apiPut(url, data, headers).then(r => r.json())
}
