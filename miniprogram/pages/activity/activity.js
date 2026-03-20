import { activityApi, registrationApi, formFieldApi, shareApi, uploadApi } from '../../utils/request'

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
    // 文件上传中状态
    uploadingFields: {},
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
      'date': '请选择日期',
      'file': '请上传文件'
    }
    return types[type] || '请输入'
  },

  // 获取文件名（从URL中提取）
  getFileName(url) {
    if (!url) return ''
    // URL可能是完整URL或相对路径
    const parts = url.split('/')
    const filename = parts[parts.length - 1]
    // 解码URL编码的文件名
    try {
      return decodeURIComponent(filename)
    } catch (e) {
      return filename
    }
  },

  // 选择文件
  chooseFile(e) {
    const fieldId = e.currentTarget.dataset.field
    wx.chooseMessageFile({
      count: 1,
      success: async (res) => {
        const tempFile = res.tempFiles[0]
        // 检查文件大小（10MB = 10 * 1024 * 1024）
        if (tempFile.size > 10 * 1024 * 1024) {
          wx.showToast({ title: '文件大小不能超过10MB', icon: 'none' })
          return
        }
        // 检查文件类型
        const ext = tempFile.name.split('.').pop().toLowerCase()
        const allowedExts = ['jpg', 'jpeg', 'png', 'gif', 'pdf', 'doc', 'docx', 'xls', 'xlsx']
        if (!allowedExts.includes(ext)) {
          wx.showToast({ title: '不支持的文件格式', icon: 'none' })
          return
        }
        // 设置上传中状态
        const { uploadingFields } = this.data
        uploadingFields[fieldId] = true
        this.setData({ uploadingFields })

        try {
          const url = await uploadApi.uploadFile(tempFile.path, tempFile.name)
          // 保存文件URL到formData
          const { formData } = this.data
          formData[fieldId] = url
          this.setData({ formData })
          wx.showToast({ title: '上传成功', icon: 'success' })
        } catch (err) {
          wx.showToast({ title: err.message || '上传失败', icon: 'none' })
        } finally {
          const { uploadingFields } = this.data
          uploadingFields[fieldId] = false
          this.setData({ uploadingFields })
        }
      },
      fail: (err) => {
        if (err.errMsg && err.errMsg.indexOf('cancel') === -1) {
          console.error('选择文件失败:', err)
          wx.showToast({ title: '选择文件失败', icon: 'none' })
        }
      }
    })
  },

  // 移除已上传的文件
  removeFile(e) {
    const fieldId = e.currentTarget.dataset.field
    const { formData } = this.data
    delete formData[fieldId]
    this.setData({ formData })
  },

  // 预览文件（图片直接预览，其他尝试下载）
  previewFile(e) {
    const url = e.currentTarget.dataset.url
    if (!url) return
    const ext = url.split('.').pop().toLowerCase()
    const imageExts = ['jpg', 'jpeg', 'png', 'gif']
    if (imageExts.includes(ext)) {
      wx.previewImage({ urls: [url] })
    } else {
      wx.showToast({ title: '点击文件可下载', icon: 'none' })
      // 复制链接到剪贴板
      wx.setClipboardData({
        data: url,
        success: () => wx.showToast({ title: '链接已复制', icon: 'success' })
      })
    }
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
          const registrationId = res.data?.id || ''
          const activityName = encodeURIComponent(activity.name || '')
          wx.redirectTo({
            url: `/pages/visitor/success/success?registrationId=${registrationId}&activityName=${activityName}`
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
