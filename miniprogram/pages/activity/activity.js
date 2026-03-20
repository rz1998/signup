import { activityApi, registrationApi, formFieldApi, shareApi } from '../../utils/request'

Page({
  data: {
    activity: null,
    loading: false,
    submitting: false,
    // 固定字段
    phone: '',
    remark: '',
    // 动态表单字段
    formFields: [],
    formData: {},
    fieldErrors: {},
    // 分享追踪
    shareId: '',
    salesId: ''
  },

  onLoad(options) {
    if (options.id) {
      // 优先检查是否有 shareId（从分享链接进入）
      if (options.shareId) {
        this.loadActivityByShareId(options.shareId)
      } else {
        this.loadActivity(options.id)
      }
    }
  },

  async loadActivityByShareId(shareId) {
    this.setData({ loading: true, shareId })
    try {
      const res = await shareApi.visit(shareId)
      if (res.success && res.data) {
        this.setData({
          activity: res.data.activity,
          salesId: res.data.salesId || '',
          shareId: shareId
        })
        // 加载表单字段
        const fields = res.data.formFields || []
        fields.sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
        this.setData({ formFields: fields })
      } else {
        wx.showToast({
          title: '分享链接无效',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('加载活动详情失败:', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadActivity(id) {
    this.setData({ loading: true })
    try {
      const res = await activityApi.get(id)
      if (res.success && res.data) {
        this.setData({
          activity: res.data
        })
        // 加载表单字段
        this.loadFormFields(id)
      } else {
        wx.showToast({
          title: '活动不存在',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('加载活动详情失败:', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadFormFields(activityId) {
    try {
      const res = await formFieldApi.list(activityId)
      if (res.success && res.data) {
        let fields = res.data.fields || []
        // Sort by sortOrder
        fields.sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
        
        // 解析选项 JSON
        fields = fields.map(field => {
          if (field.options) {
            try {
              const opts = JSON.parse(field.options)
              if (Array.isArray(opts)) {
                field.optionsArray = opts.map((opt, idx) => ({
                  label: opt,
                  value: opt
                }))
              }
            } catch (e) {
              field.optionsArray = []
            }
          } else {
            field.optionsArray = []
          }
          return field
        })
        
        this.setData({ formFields: fields })
      }
    } catch (error) {
      console.error('加载表单字段失败:', error)
    }
  },

  onPhoneInput(e) {
    this.setData({
      phone: e.detail.value
    })
  },

  onRemarkInput(e) {
    this.setData({
      remark: e.detail.value
    })
  },

  onFieldInput(e) {
    const fieldId = e.currentTarget.dataset.field
    const value = e.detail.value
    const { formData, fieldErrors } = this.data
    formData[fieldId] = value
    delete fieldErrors[fieldId]
    this.setData({ formData, fieldErrors })
  },

  onRadioChange(e) {
    const fieldId = e.currentTarget.dataset.field
    const value = e.detail.value
    const { formData, fieldErrors } = this.data
    formData[fieldId] = value
    delete fieldErrors[fieldId]
    this.setData({ formData, fieldErrors })
  },

  onCheckboxChange(e) {
    const fieldId = e.currentTarget.dataset.field
    const values = e.detail.value
    const { formData, fieldErrors } = this.data
    formData[fieldId] = values
    delete fieldErrors[fieldId]
    this.setData({ formData, fieldErrors })
  },

  onSelectChange(e) {
    const fieldId = e.currentTarget.dataset.field
    const { formFields, formData, fieldErrors } = this.data
    const field = formFields.find(f => f.id === fieldId)
    if (field && field.optionsArray) {
      formData[fieldId] = field.optionsArray[e.detail.value].value
    }
    delete fieldErrors[fieldId]
    this.setData({ formData, fieldErrors })
  },

  getFieldTypeLabel(type) {
    const types = {
      'text': '请输入',
      'textarea': '请输入',
      'number': '请输入数字',
      'tel': '请输入手机号',
      'email': '请输入邮箱',
      'radio': '请选择',
      'checkbox': '请选择',
      'select': '请选择',
      'date': '请选择日期'
    }
    return types[type] || '请输入'
  },

  async handleRegister() {
    const { activity, phone, formData, formFields, fieldErrors } = this.data
    
    if (!activity) return

    // 验证手机号
    if (!phone) {
      wx.showToast({
        title: '请输入手机号',
        icon: 'none'
      })
      return
    }

    if (!/^1[3-9]\d{9}$/.test(phone)) {
      wx.showToast({
        title: '请输入正确的手机号',
        icon: 'none'
      })
      return
    }

    // 验证动态字段
    const errors = {}
    for (const field of formFields) {
      if (field.isRequired) {
        const value = formData[field.id]
        if (!value || (typeof value === 'string' && !value.trim())) {
          errors[field.id] = `请填写${field.fieldName}`
        }
      }
    }

    if (Object.keys(errors).length > 0) {
      this.setData({ fieldErrors: errors })
      const firstError = Object.values(errors)[0]
      wx.showToast({
        title: firstError,
        icon: 'none'
      })
      return
    }

    if (activity.maxParticipants > 0 && activity.currentParticipants >= activity.maxParticipants) {
      wx.showToast({
        title: '名额已满',
        icon: 'none'
      })
      return
    }

    this.setData({ submitting: true })
    try {
      // 构建报名数据
      const registrationData = {
        activityId: activity.id,
        phone: phone
      }

      // 如果是从分享链接进入的，携带 salesId 用于追踪
      if (this.data.salesId) {
        registrationData.salesId = this.data.salesId
      }

      // 添加动态字段数据
      const extraFields = {}
      for (const field of formFields) {
        extraFields[`field_${field.id}`] = formData[field.id] || ''
        extraFields[`field_${field.id}_name`] = field.fieldName
        extraFields[`field_${field.id}_type`] = field.fieldType
      }
      registrationData.formData = JSON.stringify(extraFields)

      if (this.data.remark) {
        registrationData.remark = this.data.remark
      }

      const res = await registrationApi.create(registrationData)
      
      if (res.success) {
        wx.showToast({
          title: '报名成功',
          icon: 'success'
        })
        setTimeout(() => {
          wx.switchTab({
            url: '/pages/my-registration/my-registration'
          })
        }, 1500)
      }
    } catch (error) {
      console.error('报名失败:', error)
      wx.showToast({
        title: error.message || '报名失败',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
    }
  },

  onShareAppMessage() {
    const { activity } = this.data
    if (activity) {
      return {
        title: activity.name,
        path: `/pages/activity/activity?id=${activity.id}`
      }
    }
    return {}
  }
})
