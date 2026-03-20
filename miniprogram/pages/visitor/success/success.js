import { registrationApi } from '../../../utils/request'

Page({
  data: {
    registration: null,
    statusText: '待审核',
    formattedTime: '',
    tips: [
      '您可以点击"查看我的报名"随时查看报名状态',
      '报名信息审核通过后，我们会通过短信通知您',
      '如需修改报名信息，请在报名列表中操作',
      '如有疑问，请联系活动主办方'
    ]
  },

  onLoad(options) {
    if (options.registrationId) {
      this.loadRegistration(options.registrationId)
    }
    
    // 设置当前时间
    const now = new Date()
    const formatted = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`
    this.setData({ formattedTime: formatted })
  },

  async loadRegistration(registrationId) {
    try {
      const res = await registrationApi.get(registrationId)
      if (res.success && res.data) {
        const reg = res.data
        const statusMap = {
          'pending': '待审核',
          'confirmed': '已确认',
          'rejected': '已拒绝',
          'cancelled': '已取消'
        }
        this.setData({
          registration: reg,
          statusText: statusMap[reg.status] || reg.status
        })
      }
    } catch (error) {
      console.error('加载报名详情失败:', error)
    }
  },

  goToMyRegistration() {
    wx.switchTab({
      url: '/pages/my-registration/my-registration'
    })
  },

  goBackToActivity() {
    wx.switchTab({
      url: '/pages/index/index'
    })
  }
})
