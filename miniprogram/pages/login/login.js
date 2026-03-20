import { userApi } from '../../utils/request'

Page({
  data: {
    username: '',
    password: '',
    loading: false
  },

  onUsernameInput(e) {
    this.setData({
      username: e.detail.value
    })
  },

  onPasswordInput(e) {
    this.setData({
      password: e.detail.value
    })
  },

  async handleLogin() {
    const { username, password } = this.data

    if (!username) {
      wx.showToast({
        title: '请输入用户名',
        icon: 'none'
      })
      return
    }

    if (!password) {
      wx.showToast({
        title: '请输入密码',
        icon: 'none'
      })
      return
    }

    this.setData({ loading: true })
    try {
      const res = await userApi.login(username, password)
      console.log('Login response:', res)
      
      if (res.success && res.data && res.data.token) {
        // 保存 token 和用户信息
        wx.setStorageSync('token', res.data.token)
        wx.setStorageSync('userInfo', res.data.user)
        
        const app = getApp()
        app.globalData.token = res.data.token
        app.globalData.userInfo = res.data.user
        
        wx.showToast({
          title: '登录成功',
          icon: 'success'
        })
        
        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      } else {
        wx.showToast({
          title: res.message || res.msg || '登录失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('登录失败:', error)
      wx.showToast({
        title: error.message || '登录失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  goBack() {
    wx.navigateBack()
  }
})
