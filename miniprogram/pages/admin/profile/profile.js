import { userApi } from '../../../utils/request'

Page({
  data: {
    userInfo: null,
    formData: {},
    loading: false,
    submitting: false
  },

  onLoad() {
    this.loadUserInfo()
  },

  async loadUserInfo() {
    this.setData({ loading: true })
    try {
      const res = await userApi.getInfo()
      if (res.code === 0) {
        this.setData({
          userInfo: res.data,
          formData: { ...res.data }
        })
      }
    } catch (error) {
      console.error('加载用户信息失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({
      [`formData.${field}`]: value
    })
  },

  async submitForm() {
    const { formData } = this.data
    
    this.setData({ submitting: true })
    try {
      await userApi.update(formData)
      
      wx.setStorageSync('userInfo', formData)
      
      const app = getApp()
      app.globalData.userInfo = formData
      
      wx.showToast({
        title: '保存成功',
        icon: 'success'
      })
    } catch (error) {
      console.error('保存失败:', error)
    } finally {
      this.setData({ submitting: false })
    }
  },

  goBack() {
    wx.navigateBack()
  },

  onLogout() {
    wx.showModal({
      title: '确认退出',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          wx.removeStorageSync('token')
          wx.removeStorageSync('userInfo')
          wx.reLaunch({
            url: '/pages/login/login'
          })
        }
      }
    })
  }
})
