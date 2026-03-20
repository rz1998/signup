import { activityApi } from '../../utils/request'

Page({
  data: {
    activities: [],
    loading: false,
    current: 1
  },

  onLoad() {
    this.loadActivities()
  },

  onShow() {
    this.loadActivities()
  },

  // 日期格式化
  formatDateShort(dateStr) {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}年${month}月${day}日`
  },

  async loadActivities() {
    this.setData({ loading: true })
    try {
      const res = await activityApi.list(1, 10)
      if (res.success && res.data) {
        // Filter and format activities
        const allActivities = res.data.activities || res.data.list || []
        const activities = allActivities.filter(a => 
          a.name && a.name.trim() && a.status && a.status !== 'deleted'
        ).map(a => ({
          ...a,
          startDate: a.startDate ? this.formatDateShort(a.startDate) : '',
          endDate: a.endDate ? this.formatDateShort(a.endDate) : ''
        }))
        this.setData({
          activities: activities
        })
      }
    } catch (error) {
      console.error('加载活动失败:', error)
      wx.showToast({
        title: error.message || '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  onPullDownRefresh() {
    this.loadActivities().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  goToActivity(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/activity/activity?id=${id}`
    })
  },

  goToAdmin() {
    const token = wx.getStorageSync('token')
    if (!token) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      })
      setTimeout(() => {
        wx.navigateTo({
          url: '/pages/login/login'
        })
      }, 1500)
      return
    }
    wx.navigateTo({
      url: '/pages/admin/dashboard/dashboard'
    })
  }
})
