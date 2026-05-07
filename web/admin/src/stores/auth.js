import { ref } from 'vue'
import { adminApi } from '../api'

export const user = ref(null)
export const token = ref(localStorage.getItem('admin_token') || '')

export async function login(phone, password) {
  const resp = await adminApi.login({ phone, password })
  if (resp.error) throw new Error(resp.error)
  if (resp.user?.role !== 'admin') throw new Error('需要管理员权限')
  token.value = resp.token
  user.value = resp.user
  localStorage.setItem('admin_token', resp.token)
  return resp
}

export function logout() {
  token.value = ''
  user.value = null
  localStorage.removeItem('admin_token')
}
