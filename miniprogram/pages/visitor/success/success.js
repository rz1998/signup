import { registrationApi } from '../../../utils/request'

Page({
  data: {
    registration: null,
    statusText: '待审核',
    formattedTime: '',
    // 从URL参数传入的活动名称（作为备用显示）
    activityName: '',
    // 分享相关
    shareInfo: null,
    tips: [
      '您可以点击"查看我的报名"随时查看报名状态',
      '报名信息审核通过后，我们会通过短信通知您',
      '如需修改报名信息，请在报名列表中操作',
      '如有疑问，请联系活动主办方'
    ]
  },

  onLoad(options) {
    // 保存活动名称（从URL参数传入的中文名称）
    if (options.activityName) {
      try {
        this.setData({ activityName: decodeURIComponent(options.activityName) })
      } catch (e) {
        this.setData({ activityName: options.activityName })
      }
    }

    // 设置当前时间
    const now = new Date()
    const formatted = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')} ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`
    this.setData({ formattedTime: formatted })

    if (options.registrationId) {
      this.loadRegistration(options.registrationId)
    } else {
      // 没有 registrationId 时，尝试从本地存储的手机号获取最近一条报名
      this.loadLatestRegistration()
    }
  },

  async loadLatestRegistration() {
    const phone = wx.getStorageSync('visitorPhone')
    if (!phone) return

    try {
      const res = await registrationApi.getByPhone(phone)
      if (res.success && res.data && res.data.registrations && res.data.registrations.length > 0) {
        const reg = res.data.registrations[0]
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

  async loadRegistration(registrationId) {
    wx.showLoading({ title: '加载中...' })
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
        // 解析 formData JSON 字符串
        let formDataObj = {}
        if (reg.formData) {
          try {
            formDataObj = typeof reg.formData === 'string' ? JSON.parse(reg.formData) : reg.formData
          } catch (e) {
            console.error('formData 解析失败:', e)
          }
        }
        // 提取姓名字段（兼容 field_xxx_name 和直接字段名两种格式）
        const nameField = Object.keys(formDataObj).find(k => k.endsWith('_name') && formDataObj[k] === '姓名')
        const nameFieldKey = nameField ? nameField.replace('_name', '') : null
        const visitorName = nameFieldKey ? formDataObj[nameFieldKey] : '-'

        this.setData({
          registration: reg,
          statusText: statusMap[reg.status] || reg.status,
          // 同时保存解析后的 formData 供 wxml 使用
          _formData: formDataObj,
          _visitorName: visitorName,
          // 如果 URL 没有传 activityName，尝试从 formData 中获取
          activityName: this.data.activityName || reg.activityName || (nameFieldKey ? null : '')
        })
      } else {
        wx.showToast({ title: '未找到报名记录', icon: 'none' })
      }
    } catch (error) {
      console.error('加载报名详情失败:', error)
      wx.showToast({ title: '加载报名详情失败', icon: 'none' })
    } finally {
      wx.hideLoading()
    }
  },

  // 分享到微信
  onShareAppMessage() {
    const { registration, activityName } = this.data
    const name = activityName || registration?.activityName || '活动报名'
    return {
      title: `我已成功报名「${name}」，你也来参加吧！`,
      path: `/pages/activity/activity?id=${registration?.activityId || ''}`,
      imageUrl: '/assets/images/share-success.png'
    }
  },

  // 分享到朋友圈
  onShareTimeline() {
    const { registration, activityName } = this.data
    const name = activityName || registration?.activityName || '活动报名'
    return {
      title: `我已成功报名「${name}」，你也来参加吧！`,
      query: registration?.activityId ? `id=${registration.activityId}` : ''
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
