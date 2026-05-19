/**
 * 小程序配置
 * 生产环境替换 API_BASE 为实际域名
 */
const config = {
  // API 服务地址（不含 /api 后缀，请求时自动拼接）
  API_BASE: 'https://gjz.shadouyou.cloud/api',

  // 本地开发调试地址（仅 HBuilder 真机/模拟器可用）
  DEV_API_BASE: 'http://localhost:8080/api',
}

// 自动选择：HBuilder 真机调试用 dev，否则用生产
// 可通过 uni.getSystemInfoSync() 判断环境，暂固定生产
export default config
