import { userApi, activityApi, adminRegistrationApi } from '../../../utils/request'

Page({
  data: {
    userInfo: null,
    userRole: '',
    canManageCompanies: false,
    canManageBranches: false,
    canManageActivities: false,
    canManageUsers: false,
    canViewAllRegistrations: false,
    stats: {
      totalActivities: 0,
      totalRegistrations: 0,
      pendingReviews: 0,
      totalUsers: 0
    },
    recentActivities: [],
    recentRegistrations: [],
    loading: false
  },

  onLoad() {
    this.checkLogin()
  },

  onShow() {
    this.checkLogin()
    this.loadDashboardData()
  },

  checkLogin() {
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')
    
    if (!token) {
      wx.redirectTo({
        url: '/pages/login/login'
      })
      return
    }
    
    const app = getApp()
    const role = userInfo?.role || ''
    
    this.setData({ 
      userInfo: userInfo,
      userRole: role,
      canManageCompanies: app.isSystemAdmin(),
      canManageBranches: app.isBranchAdmin(),
      canManageActivities: app.isMgr(),
      canManageUsers: app.isAdmin(),
      canViewAllRegistrations: app.isMgr()
    })
  },

  async loadDashboardData() {
    this.setData({ loading: true })
    
    try {
      const app = getApp()
      const isMgr = app.isMgr()
      
      // 营销管理人员/管理员: 获取全部数据
      // 普通营销人员: 只看自己名下的报名
      let activitiesRes, registrationsRes
      
      const [actRes, regRes] = await Promise.all([
        activityApi.list(1, 100),
        isMgr ? adminRegistrationApi.list(1, 100) : Promise.resolve({ success: true, data: { registrations: [], pagination: { totalCount: 0 } } })
      ])
      
      activitiesRes = actRes
      registrationsRes = regRes
      
      // Parse activities
      let activities = []
      if (activitiesRes.success) {
        const activitiesData = activitiesRes.data.activities || activitiesRes.data.list || []
        activities = activitiesData.filter(a => a.name && a.name.trim())
      }
      
      // Parse registrations
      let registrations = []
      if (registrationsRes.success) {
        registrations = registrationsRes.data.list || registrationsRes.data.registrations || []
      }
      
      const pendingReviews = registrations.filter(r => r.status === 'pending').length
      
      this.setData({
        stats: {
          totalActivities: activities.length,
          totalRegistrations: registrations.length,
          pendingReviews,
          totalUsers: 0
        },
        recentActivities: activities.slice(0, 5),
        recentRegistrations: registrations.slice(0, 5)
      })
    } catch (error) {
      console.error('加载数据失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  navigateTo(e) {
    const url = e.currentTarget.dataset.url
    wx.navigateTo({ url })
  },

  goToLogin() {
    wx.redirectTo({
      url: '/pages/login/login'
    })
  },

  onLogout() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('token')
          wx.removeStorageSync('userInfo')
          this.goToLogin()
        }
      }
    })
  }
})
