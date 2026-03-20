import { adminRegistrationApi, activityApi, userApi } from '../../../utils/request'

Page({
  data: {
    canReview: false,
    canDelete: false,
    roleLabel: '报名列表',
    registrations: [],
    activities: [],
    loading: false,
    current: 1,
    total: 0,
    hasMore: true,
    filterActivityId: null,
    filterIndex: 0,
    selectedActivityTitle: ''
  },

  onLoad() {
    const app = getApp()
    const isMgr = app.isMgr()
    const isSales = app.isSales()
    const isAdmin = app.isAdmin()
    
    this.setData({
      canReview: isMgr,
      canDelete: isMgr,
      roleLabel: isSales ? '我的报名' : isMgr ? '报名管理' : '报名列表'
    })
    
    this.loadActivities()
    this.loadRegistrations()
  },

  async loadActivities() {
    try {
      const res = await activityApi.list(1, 100)
      if (res.success && res.data) {
        this.setData({
          activities: res.data.activities || res.data.list || []
        })
      }
    } catch (error) {
      console.error('加载活动列表失败:', error)
    }
  },

  async loadRegistrations(loadMore = false) {
    const { current, registrations, filterActivityId } = this.data
    
    this.setData({ loading: true })
    try {
      const res = await adminRegistrationApi.list(current, 10, filterActivityId || undefined)
      if (res.success && res.data) {
        const list = res.data.registrations || []
        this.setData({
          registrations: loadMore ? [...registrations, ...list] : list,
          total: res.data.pagination?.totalCount || 0,
          hasMore: list.length >= 10
        })
      }
    } catch (error) {
      console.error('加载报名列表失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({
        current: this.data.current + 1
      })
      this.loadRegistrations(true)
    }
  },

  onFilterChange(e) {
    const value = e.detail.value
    const activities = this.data.activities
    const activity = activities[value]
    const activityId = activity ? activity.id : null
    const title = activity ? activity.title : ''
    
    this.setData({
      filterActivityId: activityId,
      filterIndex: value,
      selectedActivityTitle: title,
      current: 1,
      registrations: []
    })
    this.loadRegistrations()
  },

  async reviewRegistration(e) {
    if (!this.data.canReview) {
      wx.showToast({ title: '无权限', icon: 'none' })
      return
    }
    const { id, status } = e.currentTarget.dataset
    
    try {
      await adminRegistrationApi.review(id, status)
      wx.showToast({
        title: status === 'confirmed' ? '已确认' : '已拒绝',
        icon: 'success'
      })
      this.setData({
        current: 1,
        registrations: []
      })
      this.loadRegistrations()
    } catch (error) {
      console.error('审核报名失败:', error)
    }
  },

  async deleteRegistration(e) {
    if (!this.data.canDelete) {
      wx.showToast({ title: '无权限', icon: 'none' })
      return
    }
    const id = e.currentTarget.dataset.id
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该报名记录吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await adminRegistrationApi.delete(id)
            wx.showToast({
              title: '删除成功',
              icon: 'success'
            })
            this.setData({
              current: 1,
              registrations: []
            })
            this.loadRegistrations()
          } catch (error) {
            console.error('删除报名失败:', error)
          }
        }
      }
    })
  },

  goBack() {
    wx.navigateBack()
  }
})
