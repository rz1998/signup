import { registrationApi, activityApi, formFieldApi } from '../../../utils/request'

Page({
  data: {
    registrationId: '',
    registration: null,
    activity: null,
    loading: false,
    submitting: false,
    statusText: '待审核',
    // 表单字段
    formFields: [],
    customFields: [],
    nameField: null,
    phoneField: null,
    // 表单数据
    formData: {},
    remark: '',
    fieldErrors: {}
  },

  onLoad(options) {
    if (options.registrationId) {
      this.setData({ registrationId: options.registrationId })
      this.loadRegistration(options.registrationId)
    } else {
      wx.showToast({
        title: '参数错误',
        icon: 'none'
      })
      setTimeout(() => {
        wx.navigateBack()
      }, 1500)
    }
  },

  async loadRegistration(registrationId) {
    this.setData({ loading: true })
    try {
      const res = await registrationApi.get(registrationId)
      if (res.success && res.data) {
        const reg = res.data
        const statusMap = {
          'pending': '待审核',
          'confirmed': '已确认',
          'rejected': '已拒绝',
          'cancelled': '已取消'
        }
        
        // 解析表单数据
        let formData = {}
        if (reg.formData) {
          if (typeof reg.formData === 'string') {
            try {
              formData = JSON.parse(reg.formData)
            } catch (e) {
              formData = reg.formData
            }
          } else {
            formData = { ...reg.formData }
          }
        }
        formData._remark = reg.remark || ''
        
        this.setData({
          registration: reg,
          statusText: statusMap[reg.status] || reg.status,
          formData: formData,
          remark: reg.remark || ''
        })
        
        // 加载活动信息和表单字段
        if (reg.activityId) {
          this.loadActivity(reg.activityId)
        }
      } else {
        wx.showToast({
          title: '报名信息不存在',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('加载报名详情失败:', error)
      wx.showToast({
        title: '加载失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadActivity(activityId) {
    try {
      const [activityRes, fieldsRes] = await Promise.all([
        activityApi.get(activityId),
        formFieldApi.list(activityId)
      ])
      
      if (activityRes.success && activityRes.data) {
        this.setData({ activity: activityRes.data })
      }
      
      if (fieldsRes.success && fieldsRes.data) {
        let fields = fieldsRes.data.fields || []
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
        
        // 分离默认字段和自定义字段
        const defaultFields = fields.filter(f => f.isDefault)
        const customFields = fields.filter(f => !f.isDefault)
        
        const nameField = defaultFields.find(f => f.fieldName === '姓名')
        const phoneField = defaultFields.find(f => f.fieldName === '手机号')
        
        this.setData({
          formFields: fields,
          customFields,
          nameField: nameField || null,
          phoneField: phoneField || null
        })
      }
    } catch (error) {
      console.error('加载活动信息失败:', error)
    }
  },

  onNameInput(e) {
    const { formData } = this.data
    formData.name = e.detail.value
    this.setData({ formData })
  },

  onPhoneInput(e) {
    const { formData } = this.data
    formData.phone = e.detail.value
    this.setData({ formData })
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
    const { customFields, formData, fieldErrors } = this.data
    const field = customFields.find(f => f.id === fieldId)
    if (field && field.optionsArray) {
      formData[fieldId] = field.optionsArray[e.detail.value].value
    }
    delete fieldErrors[fieldId]
    this.setData({ formData, fieldErrors })
  },

  onRemarkInput(e) {
    const { formData } = this.data
    formData._remark = e.detail.value
    this.setData({ formData })
  },

  getPlaceholder(type) {
    const placeholders = {
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
    return placeholders[type] || '请输入'
  },

  async handleSubmit() {
    const { registration, formData, formFields, fieldErrors, nameField, phoneField } = this.data
    
    if (!registration) return
    
    // 验证姓名
    if (nameField && nameField.isRequired) {
      if (!formData.name || !formData.name.trim()) {
        wx.showToast({
          title: '请输入姓名',
          icon: 'none'
        })
        return
      }
    }
    
    // 验证手机号
    if (phoneField && phoneField.isRequired) {
      if (!formData.phone) {
        wx.showToast({
          title: '请输入手机号',
          icon: 'none'
        })
        return
      }
      if (!/^1[3-9]\d{9}$/.test(formData.phone)) {
        wx.showToast({
          title: '请输入正确的手机号',
          icon: 'none'
        })
        return
      }
    }
    
    // 验证自定义字段
    const errors = {}
    for (const field of formFields) {
      if (!field.isDefault && field.isRequired) {
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

    this.setData({ submitting: true })
    try {
      // 构建提交数据
      const submitData = {
        phone: formData.phone
      }
      
      // 添加默认字段数据
      if (nameField) {
        submitData.name = formData.name
      }
      
      // 添加自定义字段数据
      const extraFields = {}
      for (const field of formFields) {
        if (!field.isDefault) {
          extraFields[`field_${field.id}`] = formData[field.id] || ''
          extraFields[`field_${field.id}_name`] = field.fieldName
          extraFields[`field_${field.id}_type`] = field.fieldType
        }
      }
      submitData.formData = JSON.stringify(extraFields)
      submitData.remark = formData._remark || ''

      const res = await registrationApi.updateByVisitor(registration.id, submitData)
      
      if (res.success) {
        wx.showToast({
          title: '修改成功',
          icon: 'success'
        })
        setTimeout(() => {
          wx.navigateBack()
        }, 1500)
      } else {
        wx.showToast({
          title: res.message || '修改失败',
          icon: 'none'
        })
      }
    } catch (error) {
      console.error('修改报名失败:', error)
      wx.showToast({
        title: error.message || '修改失败',
        icon: 'none'
      })
    } finally {
      this.setData({ submitting: false })
    }
  },

  goBack() {
    wx.navigateBack()
  }
})
