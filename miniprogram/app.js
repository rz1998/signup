// app.js
App({
  globalData: {
    userInfo: null,
    token: '',
    apiBaseUrl: 'http://192.168.8.230:8082/api/v1'
  },
  onLaunch() {
    // 校验 token 是否存在
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
    }
    
    // 获取用户信息
    const userInfo = wx.getStorageSync('userInfo')
    if (userInfo) {
      this.globalData.userInfo = userInfo
    }
  },
  onShow() {
    // 小程序显示时检查登录状态
  },
  onHide() {
    // 小程序隐藏时
  },
  // 权限判断方法
  hasPermission(...roles) {
    const userInfo = this.globalData.userInfo
    if (!userInfo) return false
    return roles.includes(userInfo.role)
  },
  isAdmin() {
    return this.hasPermission('admin')
  },
  isSystemAdmin() {
    return this.hasPermission('admin')
  },
  isCompanyAdmin() {
    return this.hasPermission('admin', 'company_mgr')
  },
  isBranchAdmin() {
    return this.hasPermission('admin', 'company_mgr', 'branch_mgr')
  },
  isMgr() {
    return this.hasPermission('admin', 'company_mgr', 'branch_mgr', 'mgr')
  },
  isSales() {
    return this.hasPermission('sales')
  }
})
