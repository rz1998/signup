import { activityApi, formFieldApi } from '../../../utils/request'

Page({
  data: {
    activityId: '',
    activityTitle: '',
    loading: false,
    // Tab
    activeTab: 'fields',
    // Form fields
    fields: [],
    showFieldForm: false,
    fieldFormData: {
      id: '',
      name: '',
      label: '',
      type: 'text',
      required: false,
      placeholder: '',
      options: '',
      sortOrder: 0
    },
    isEditField: false,
    // Field type options
    fieldTypes: [
      { value: 'text', label: '文本输入' },
      { value: 'number', label: '数字输入' },
      { value: 'select', label: '单选下拉' },
      { value: 'multi_select', label: '多选下拉' },
      { value: 'date', label: '日期选择' },
      { value: 'time', label: '时间选择' },
      { value: 'phone', label: '手机号' },
      { value: 'email', label: '邮箱' }
    ],
    selectedFieldTypeLabel: '文本输入',
    // Page style config
    styleData: {
      pageTitle: '',
      pageSubtitle: '',
      pageDescription: '',
      coverImage: '',
      backgroundColor: '#f5f5f5'
    },
    showStyleForm: false
  },

  onLoad(options) {
    const { activityId, activityTitle } = options
    if (!activityId) {
      wx.showToast({ title: '参数错误', icon: 'none' })
      wx.navigateBack()
      return
    }
    this.setData({
      activityId,
      activityTitle: decodeURIComponent(activityTitle || '')
    })
    this.loadFields()
    this.loadStyleConfig()
  },

  goBack() {
    wx.navigateBack()
  },

  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({ activeTab: tab })
  },

  async loadFields() {
    const { activityId } = this.data
    this.setData({ loading: true })
    try {
      const res = await formFieldApi.list(activityId)
      if (res.success && res.data) {
        const fields = (res.data.fields || []).map((f, index) => ({
          id: f.id,
          name: f.name || '',
          label: f.label || '',
          type: f.type || 'text',
          required: f.required || false,
          placeholder: f.placeholder || '',
          options: Array.isArray(f.options) ? f.options.join('\n') : (f.options || ''),
          sortOrder: f.sortOrder ?? index
        }))
        // Sort by sortOrder
        fields.sort((a, b) => a.sortOrder - b.sortOrder)
        this.setData({ fields })
      }
    } catch (error) {
      console.error('加载字段失败:', error)
      wx.showToast({ title: '加载字段失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async loadStyleConfig() {
    const { activityId } = this.data
    try {
      const res = await activityApi.get(activityId)
      if (res.success && res.data) {
        const act = res.data
        this.setData({
          styleData: {
            pageTitle: act.pageTitle || '',
            pageSubtitle: act.pageSubtitle || '',
            pageDescription: act.pageDescription || '',
            coverImage: act.coverImage || '',
            backgroundColor: act.backgroundColor || '#f5f5f5'
          }
        })
      }
    } catch (error) {
      console.error('加载页面配置失败:', error)
    }
  },

  // ==================== 字段管理 ====================
  showAddField() {
    this.setData({
      showFieldForm: true,
      fieldFormData: {
        id: '',
        name: '',
        label: '',
        type: 'text',
        required: false,
        placeholder: '',
        options: '',
        sortOrder: 0
      },
      selectedFieldTypeLabel: '文本输入',
      isEditField: false
    })
  },

  showEditField(e) {
    const field = e.currentTarget.dataset.field
    const fieldType = this.data.fieldTypes.find(t => t.value === field.type)
    const fieldTypeLabel = fieldType ? fieldType.label : '文本输入'
    this.setData({
      showFieldForm: true,
      fieldFormData: {
        id: field.id,
        name: field.name,
        label: field.label,
        type: field.type,
        required: field.required,
        placeholder: field.placeholder || '',
        options: field.options || '',
        sortOrder: field.sortOrder
      },
      selectedFieldTypeLabel: fieldTypeLabel,
      isEditField: true
    })
  },

  hideFieldForm() {
    this.setData({ showFieldForm: false })
  },

  onFieldFormInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`fieldFormData.${field}`]: value })
  },

  onFieldTypeChange(e) {
    const selectedType = this.data.fieldTypes[e.detail.value]
    this.setData({
      'fieldFormData.type': selectedType.value,
      selectedFieldTypeLabel: selectedType.label
    })
  },

  onFieldRequiredChange(e) {
    this.setData({ 'fieldFormData.required': e.detail.value })
  },

  async submitFieldForm() {
    const { fieldFormData, isEditField, activityId } = this.data

    if (!fieldFormData.label) {
      wx.showToast({ title: '请填写字段标签', icon: 'none' })
      return
    }
    if (!fieldFormData.name) {
      wx.showToast({ title: '请填写字段名称', icon: 'none' })
      return
    }

    const apiData = {
      name: fieldFormData.name,
      label: fieldFormData.label,
      type: fieldFormData.type,
      required: fieldFormData.required,
      placeholder: fieldFormData.placeholder,
      options: fieldFormData.type === 'select' || fieldFormData.type === 'multi_select'
        ? fieldFormData.options.split('\n').filter(o => o.trim())
        : []
    }

    this.setData({ loading: true })
    try {
      if (isEditField && fieldFormData.id) {
        await formFieldApi.update(fieldFormData.id, apiData)
        wx.showToast({ title: '更新成功', icon: 'success' })
      } else {
        await formFieldApi.create(activityId, apiData)
        wx.showToast({ title: '创建成功', icon: 'success' })
      }
      this.hideFieldForm()
      this.loadFields()
    } catch (error) {
      console.error('保存字段失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async deleteField(e) {
    const field = e.currentTarget.dataset.field
    wx.showModal({
      title: '确认删除',
      content: `确定要删除字段「${field.label}」吗？`,
      success: async (res) => {
        if (res.confirm) {
          try {
            await formFieldApi.delete(field.id)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.loadFields()
          } catch (error) {
            console.error('删除字段失败:', error)
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  // Move field up
  moveFieldUp(e) {
    const index = e.currentTarget.dataset.index
    if (index <= 0) return
    const fields = [...this.data.fields]
    const temp = fields[index]
    fields[index] = fields[index - 1]
    fields[index - 1] = temp
    this.setData({ fields })
    this.saveFieldOrder(fields)
  },

  // Move field down
  moveFieldDown(e) {
    const index = e.currentTarget.dataset.index
    const fields = [...this.data.fields]
    if (index >= fields.length - 1) return
    const temp = fields[index]
    fields[index] = fields[index + 1]
    fields[index + 1] = temp
    this.setData({ fields })
    this.saveFieldOrder(fields)
  },

  async saveFieldOrder(fields) {
    const { activityId } = this.data
    const fieldIds = fields.map(f => f.id)
    try {
      // Sort fields by updated order and assign sortOrder
      const sortedFields = fields.map((f, idx) => ({ ...f, sortOrder: idx }))
      await formFieldApi.reorder(activityId, fieldIds)
    } catch (error) {
      console.error('保存排序失败:', error)
    }
  },

  // ==================== 页面样式配置 ====================
  showStyleConfig() {
    this.setData({ showStyleForm: true })
  },

  hideStyleForm() {
    this.setData({ showStyleForm: false })
  },

  onStyleInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`styleData.${field}`]: value })
  },

  chooseCoverImage() {
    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const filePath = res.tempFiles[0].tempFilePath
        wx.uploadFile({
          url: `${getApp().globalData.apiBaseUrl || 'http://192.168.8.230:8082/api/v1'}/upload`,
          filePath,
          name: 'file',
          header: {
            'Authorization': `Bearer ${wx.getStorageSync('token')}`
          },
          success: (uploadRes) => {
            try {
              const data = JSON.parse(uploadRes.data)
              if (data.success && data.data && data.data.url) {
                this.setData({ 'styleData.coverImage': data.data.url })
                wx.showToast({ title: '上传成功', icon: 'success' })
              } else {
                wx.showToast({ title: '上传失败', icon: 'none' })
              }
            } catch (e) {
              wx.showToast({ title: '上传失败', icon: 'none' })
            }
          },
          fail: () => {
            wx.showToast({ title: '上传失败', icon: 'none' })
          }
        })
      }
    })
  },

  async saveStyleConfig() {
    const { styleData, activityId } = this.data
    this.setData({ loading: true })
    try {
      await activityApi.update(activityId, {
        pageTitle: styleData.pageTitle,
        pageSubtitle: styleData.pageSubtitle,
        pageDescription: styleData.pageDescription,
        coverImage: styleData.coverImage,
        backgroundColor: styleData.backgroundColor
      })
      wx.showToast({ title: '保存成功', icon: 'success' })
      this.hideStyleForm()
    } catch (error) {
      console.error('保存页面配置失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  getFieldTypeLabel(type) {
    const found = this.data.fieldTypes.find(t => t.value === type)
    return found ? found.label : type
  }
})
