const api = require('../../utils/api.js')

const PRODUCT_NAMES = { single: '单次测算', monthly: '月卡', quarterly: '季卡', yearly: '年卡' }
const STATUS_MAP = { pending: '待支付', paid: '已支付', failed: '支付失败', refunded: '已退款' }

Page({
  data: { orders: [], loading: true },

  onShow() { this.load() },

  load() {
    this.setData({ loading: true })
    api.listOrders().then(d => {
      const orders = (d.items || []).map(o => ({
        id: o.id,
        name: PRODUCT_NAMES[o.product_id] || (o.product_id || '订单'),
        amount: (o.amount / 100).toFixed(2),
        status: STATUS_MAP[o.status] || o.status,
        isPaid: o.status === 'paid',
        time: (o.paid_at || o.created_at || '').replace('T', ' ').slice(0, 16)
      }))
      this.setData({ orders, loading: false })
    }).catch(() => this.setData({ loading: false }))
  },

  goRecharge() { wx.navigateTo({ url: '/pages/user/user' }) },
  goHome() { wx.reLaunch({ url: '/pages/index/index' }) }
})
