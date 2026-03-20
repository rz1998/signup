import { settingsApi, activityApi, formFieldApi } from '../../../utils/request'

Page({
  data: {
    settings: {
      homeTitle: '活动报名',
      homeSubtitle: '欢迎参加我们的活动',
      homeDescription: '这里汇集了各类精彩活动，等你来参与',
      registrationTitle: '活动报名',
      registrationSuccessMsg: '恭喜您报名成功！',
      registrationNotice: '请确保填写真实信息',
      bgColor: '#FFFFFF',
      bgImage: '',
      contactPhone: '',
      contactEmail: ''
    },
    // Editor states
    settingsData: {
      pageTitle: '',
      pageSubtitle: '',
      pageDescription: '',
      registrationNotice: '',
      contactPhone: '',
      contactEmail: ''
    },
    activityId: '',
    activityTitle: '',
    loading: false,
    
    // 表单字段管理
    formFields: [],
    showFormFields: false,
    showFieldForm: false,
    editingField: null,
    fieldFormData: {
      fieldName: '',
      fieldType: 'text',
      isRequired: false
    },
    fieldTypes: [
      { value: 'text', label: '单行文本' },
      { value: 'textarea', label: '多行文本' },
      { value: 'number', label: '数字' },
      { value: 'tel', label: '手机号' },
      { value: 'email', label: '邮箱' },
      { value: 'radio', label: '单选' },
      { value: 'checkbox', label: '多选' },
      { value: 'select', label: '下拉选择' },
      { value: 'date', label: '日期' }
    ],
    fieldTypeLabel: '单行文本'
  },

  onLoad(options) {
    if (options.activityId) {
      this.setData({
        activityId: options.activityId,
        activityTitle: options.activityTitle || ''
      })
      this.loadActivitySettings()
    } else {
      this.loadGlobalSettings()
    }
  },

  async loadGlobalSettings() {
    this.setData({ loading: true })
    try {
      const res = await settingsApi.get()
      if (res.success && res.data) {
        this.setData({
          settings: { ...this.data.settings, ...res.data }
        })
      }
    } catch (error) {
      console.error('加载设置失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadActivitySettings() {
    this.setData({ loading: true })
    try {
      // Load activity details for page settings
      const res = await activityApi.get(this.data.activityId)
      if (res.success && res.data) {
        const activity = res.data
        this.setData({
          settingsData: {
            pageTitle: activity.pageTitle || '',
            pageSubtitle: activity.pageSubtitle || '',
            pageDescription: activity.pageDescription || '',
            registrationNotice: activity.registrationNotice || '',
            contactPhone: activity.contactPhone || '',
            contactEmail: activity.contactEmail || ''
          }
        })
      }
    } catch (error) {
      console.error('加载活动设置失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  // 加载表单字段
  async loadFormFields() {
    if (!this.data.activityId) return
    
    try {
      const res = await formFieldApi.list(this.data.activityId)
      if (res.success && res.data) {
        const fields = res.data.fields || []
        // Sort by sortOrder
        fields.sort((a, b) => (a.sortOrder || 0) - (b.sortOrder || 0))
        this.setData({ formFields: fields })
      }
    } catch (error) {
      console.error('加载表单字段失败:', error)
    }
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({
      [`settingsData.${field}`]: value
    })
  },

  // 插入富文本标签
  insertText(e) {
    const tag = e.currentTarget.dataset.tag
    const currentContent = this.data.settingsData.pageDescription || ''
    let insert = ''
    
    switch(tag) {
      case '加粗':
        insert = '**加粗文字**'
        break
      case '斜体':
        insert = '*斜体文字*'
        break
      case '链接':
        insert = '[链接文字](https://)'
        break
      case '图片':
        insert = '![图片描述](图片URL)'
        break
      case '标题1':
        insert = '\n# 标题1\n'
        break
      case '标题2':
        insert = '\n## 标题2\n'
        break
      case '引用':
        insert = '\n> 引用内容\n'
        break
      case '列表':
        insert = '\n- 列表项1\n- 列表项2\n'
        break
      default:
        insert = tag
    }
    
    this.setData({
      'settingsData.pageDescription': currentContent + insert
    })
  },

  async saveSettings() {
    const { settingsData, activityId } = this.data
    
    this.setData({ loading: true })
    try {
      if (activityId) {
        // Save to activity
        await activityApi.update(activityId, {
          pageTitle: settingsData.pageTitle,
          pageSubtitle: settingsData.pageSubtitle,
          pageDescription: settingsData.pageDescription,
          registrationNotice: settingsData.registrationNotice,
          contactPhone: settingsData.contactPhone,
          contactEmail: settingsData.contactEmail
        })
      } else {
        // Save global settings
        await settingsApi.save(this.data.settings)
      }
      
      wx.showToast({
        title: '保存成功',
        icon: 'success',
        duration: 1000
      })
      
      // Save successful, go back after toast
      setTimeout(() => {
        wx.navigateBack()
      }, 1000)
    } catch (error) {
      console.error('保存设置失败:', error)
      wx.showToast({
        title: '保存失败: ' + (error.message || '网络错误'),
        icon: 'none',
        duration: 2000
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  goBack() {
    wx.navigateBack()
  },

  // ========== 表单字段管理 ==========

  toggleFormFields() {
    const show = !this.data.showFormFields
    this.setData({ showFormFields: show })
    if (show && this.data.activityId) {
      this.loadFormFields()
    }
  },

  showAddFieldForm() {
    this.setData({
      showFieldForm: true,
      editingField: null,
      fieldFormData: {
        fieldName: '',
        fieldType: 'text',
        isRequired: false,
        options: ''  // 选项，用换行分隔
      },
      fieldTypeLabel: '单行文本'
    })
  },

  showEditFieldForm(e) {
    const field = e.currentTarget.dataset.field
    // 解析 options JSON
    let options = ''
    if (field.options) {
      try {
        const opts = JSON.parse(field.options)
        if (Array.isArray(opts)) {
          options = opts.join('\n')
        }
      } catch (e) {
        options = field.options || ''
      }
    }
    // 计算 fieldTypeLabel
    const fieldTypes = this.data.fieldTypes
    const ft = fieldTypes.find(t => t.value === field.fieldType)
    const fieldTypeLabel = ft ? ft.label : '单行文本'
    this.setData({
      showFieldForm: true,
      editingField: field,
      fieldFormData: {
        fieldName: field.fieldName,
        fieldType: field.fieldType,
        isRequired: field.isRequired,
        options: options
      },
      fieldTypeLabel: fieldTypeLabel
    })
  },

  hideFieldForm() {
    this.setData({
      showFieldForm: false,
      editingField: null
    })
  },

  onFieldInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({
      [`fieldFormData.${field}`]: value
    })
  },

  onFieldTypeChange(e) {
    const fieldTypes = this.data.fieldTypes
    const selectedIndex = parseInt(e.detail.value, 10)
    const selectedType = fieldTypes[selectedIndex]
    this.setData({
      'fieldFormData.fieldType': selectedType.value,
      fieldTypeLabel: selectedType.label
    })
  },

  onFieldRequiredChange(e) {
    this.setData({
      'fieldFormData.isRequired': e.detail.value.length > 0
    })
  },

  async saveField() {
    const { fieldFormData, activityId, editingField } = this.data
    
    if (!fieldFormData.fieldName) {
      wx.showToast({ title: '请输入字段名称', icon: 'none' })
      return
    }

    // 检查需要选项的字段类型
    const needsOptions = ['radio', 'checkbox', 'select']
    if (needsOptions.includes(fieldFormData.fieldType) && !fieldFormData.options.trim()) {
      wx.showToast({ title: '请输入选项内容', icon: 'none' })
      return
    }

    // 将选项转换为JSON数组
    let optionsJson = ''
    if (fieldFormData.options.trim()) {
      const optionsArray = fieldFormData.options.split('\n').filter(o => o.trim())
      optionsJson = JSON.stringify(optionsArray)
    }

    try {
      if (editingField) {
        // Update existing field
        await formFieldApi.update(editingField.id, {
          fieldName: fieldFormData.fieldName,
          fieldType: fieldFormData.fieldType,
          isRequired: fieldFormData.isRequired,
          options: optionsJson
        })
        wx.showToast({ title: '更新成功', icon: 'success' })
      } else {
        // Create new field
        await formFieldApi.create(activityId, {
          fieldName: fieldFormData.fieldName,
          fieldType: fieldFormData.fieldType,
          isRequired: fieldFormData.isRequired,
          options: optionsJson,
          sortOrder: this.data.formFields.length
        })
        wx.showToast({ title: '添加成功', icon: 'success' })
      }
      
      this.hideFieldForm()
      this.loadFormFields()
    } catch (error) {
      console.error('保存字段失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    }
  },

  async deleteField(e) {
    const fieldId = e.currentTarget.dataset.id
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该字段吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await formFieldApi.delete(fieldId)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.loadFormFields()
          } catch (error) {
            console.error('删除字段失败:', error)
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  async moveField(e) {
    const { id, direction } = e.currentTarget.dataset
    const fields = [...this.data.formFields]
    const index = fields.findIndex(f => f.id === id)
    
    if (direction === 'up' && index > 0) {
      [fields[index - 1], fields[index]] = [fields[index], fields[index - 1]]
    } else if (direction === 'down' && index < fields.length - 1) {
      [fields[index], fields[index + 1]] = [fields[index + 1], fields[index]]
    } else {
      return
    }

    this.setData({ formFields: fields })

    // Update sort order
    try {
      for (let i = 0; i < fields.length; i++) {
        await formFieldApi.update(fields[i].id, {
          sortOrder: i
        })
      }
    } catch (error) {
      console.error('更新排序失败:', error)
    }
  },

  getFieldTypeLabel(type) {
    const types = {
      'text': '单行文本',
      'textarea': '多行文本',
      'number': '数字',
      'tel': '手机号',
      'email': '邮箱',
      'radio': '单选',
      'checkbox': '多选',
      'select': '下拉选择',
      'date': '日期'
    }
    return types[type] || type
  }
})
