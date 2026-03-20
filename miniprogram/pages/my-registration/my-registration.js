import { registrationApi, activityApi, formFieldApi } from '../../utils/request'

Page({
  data: {
    registrations: [],
    loading: false,
    showPhoneSearch: true,
    phone: '',
    showEditForm: false,
    editingRegistration: null,
    formData: {},
    formFields: []
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

  async showEditForm(e) {
    const reg = e.currentTarget.dataset.reg
    if (reg.status !== 'pending') {
      wx.showToast({
        title: '当前状态不可修改',
        icon: 'none'
      })
      return
    }

    // 加载表单字段
    try {
      const fieldsRes = await formFieldApi.list(reg.activityId)
      let formFields = []
      if (fieldsRes.success && fieldsRes.data) {
        formFields = fieldsRes.data.fields || []
        formFields.sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
      }

      // 解析已有数据
      const formData = reg.formData || {}
      formData._remark = reg.remark || ''

      this.setData({
        showEditForm: true,
        editingRegistration: reg,
        formFields,
        formData
      })
    } catch (error) {
      console.error('加载表单字段失败:', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    }
  },

  hideEditForm() {
    this.setData({
      showEditForm: false,
      editingRegistration: null,
      formFields: [],
      formData: {}
    })
  },

  onFieldInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    const { formData } = this.data
    formData[field] = value
    this.setData({ formData })
  },

  onRemarkInput(e) {
    const { formData } = this.data
    formData._remark = e.detail.value
    this.setData({ formData })
  },

  async submitEdit() {
    const { editingRegistration, formData, formFields, phone } = this.data

    // 验证必填字段
    for (const field of formFields) {
      if (field.isRequired && !formData[field.id]) {
        wx.showToast({
          title: `请填写${field.fieldName}`,
          icon: 'none'
        })
        return
      }
    }

    this.setData({ loading: true })
    try {
      // 构建提交数据
      const submitData = {
        phone: phone
      }

      // 添加表单字段数据
      const extraFields = {}
      for (const field of formFields) {
        extraFields[`field_${field.id}`] = formData[field.id] || ''
        extraFields[`field_${field.id}_name`] = field.fieldName
        extraFields[`field_${field.id}_type`] = field.fieldType
      }
      submitData.formData = JSON.stringify(extraFields)
      submitData.remark = formData._remark || ''

      const res = await registrationApi.updateByVisitor(editingRegistration.id, submitData)
      if (res.success) {
        wx.showToast({
          title: '修改成功',
          icon: 'success'
        })
        this.hideEditForm()
        this.loadRegistrations(phone)
      } else {
        wx.showToast({
          title: res.message || '修改失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('修改报名失败:', error)
      wx.showToast({
        title: '修改失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
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
