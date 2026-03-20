import { registrationApi } from '../../utils/request'

Page({
  data: {
    registrations: [],
    loading: false,
    showPhoneSearch: true,
    phone: ''
  },

  onLoad() {
    // Check if phone is stored locally
    const phone = wx.getStorageSync('visitorPhone')
    if (phone) {
      this.setData({ phone, showPhoneSearch: false })
      this.loadRegistrations(phone)
    }
  },

  onPhoneInput(e) {
    this.setData({
      phone: e.detail.value
    })
  },

  async searchByPhone() {
    const { phone } = this.data
    
    if (!phone || !/^1[3-9]\d{9}$/.test(phone)) {
      wx.showToast({
        title: '请输入正确的手机号',
        icon: 'none'
      })
      return
    }

    this.setData({ loading: true })
    try {
      const res = await registrationApi.getByPhone(phone)
      if (res.success && res.data) {
        // Store phone locally
        wx.setStorageSync('visitorPhone', phone)
        
        this.setData({
          registrations: res.data.registrations || [],
          showPhoneSearch: false
        })
        
        if (res.data.registrations && res.data.registrations.length === 0) {
          wx.showToast({
            title: '未找到报名记录',
            icon: 'none'
          })
        }
      } else {
        wx.showToast({
          title: '查询失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('查询报名记录失败:', error)
      wx.showToast({
        title: '查询失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadRegistrations(phone) {
    this.setData({ loading: true })
    try {
      const res = await registrationApi.getByPhone(phone)
      if (res.success && res.data) {
        this.setData({
          registrations: res.data.registrations || []
        })
      }
    } catch (error) {
      console.error('加载报名记录失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  changePhone() {
    wx.removeStorageSync('visitorPhone')
    this.setData({
      showPhoneSearch: true,
      registrations: [],
      phone: ''
    })
  },

  goToEditPage(e) {
    const reg = e.currentTarget.dataset.reg
    if (reg.status !== 'pending') {
      wx.showToast({
        title: '当前状态不可修改',
        icon: 'none'
      })
      return
    }
    wx.navigateTo({
      url: `/pages/visitor/edit/edit?registrationId=${reg.id}`
    })
  },

  async cancelRegistration(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认取消',
      content: '确定要取消该报名吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await registrationApi.cancel(id)
            wx.showToast({
              title: '已取消',
              icon: 'success'
            })
            this.loadRegistrations(this.data.phone)
          } catch (error) {
            console.error('取消报名失败:', error)
            wx.showToast({
              title: '取消失败',
              icon: 'none'
            })
          }
        }
      }
    })
  },

  getStatusText(status) {
    const statusMap = {
      'pending': '待审核',
      'confirmed': '已确认',
      'rejected': '已拒绝',
      'cancelled': '已取消'
    }
    return statusMap[status] || status
  }
})
