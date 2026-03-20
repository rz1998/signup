import { companyApi } from '../../../utils/request'

Page({
  data: {
    companies: [],
    loading: false,
    current: 1,
    total: 0,
    hasMore: true,
    keyword: '',
    showForm: false,
    formData: {},
    isEdit: false,
    statusIndex: 0,
    statusLabel: '正常',
    statuses: [
      { label: '正常', value: 'active' },
      { label: '禁用', value: 'inactive' }
    ]
  },

  onLoad() {
    const app = getApp()
    if (!app.isAdmin()) {
      wx.showToast({ title: '无权限访问', icon: 'none' })
      wx.navigateBack()
      return
    }
    this.loadCompanies()
  },

  async loadCompanies(loadMore = false) {
    const { current, companies, keyword } = this.data
    this.setData({ loading: true })
    try {
      const res = await companyApi.list(current, 20, keyword)
      if (res.success && res.data) {
        const list = res.data.companies || []
        this.setData({
          companies: loadMore ? [...companies, ...list] : list,
          total: res.data.pagination?.totalCount || 0,
          hasMore: list.length >= 20
        })
      }
    } catch (error) {
      console.error('加载公司列表失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ current: this.data.current + 1 })
      this.loadCompanies(true)
    }
  },

  onSearch(e) {
    const keyword = e.detail.value
    this.setData({ keyword, current: 1, companies: [] })
    this.loadCompanies()
  },

  showAddForm() {
    this.setData({
      showForm: true,
      formData: {},
      isEdit: false,
      statusIndex: 0,
      statusLabel: '正常'
    })
  },

  showEditForm(e) {
    const company = e.currentTarget.dataset.company
    const statusIndex = company.status === 'inactive' ? 1 : 0
    this.setData({
      showForm: true,
      formData: {
        id: company.id,
        name: company.name || '',
        description: company.description || ''
      },
      isEdit: true,
      statusIndex,
      statusLabel: company.status === 'inactive' ? '禁用' : '正常'
    })
  },

  hideForm() {
    this.setData({ showForm: false, formData: {} })
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`formData.${field}`]: value })
  },

  onStatusChange(e) {
    const index = parseInt(e.detail.value, 10)
    const status = this.data.statuses[index]
    this.setData({
      statusIndex: index,
      statusLabel: status.label,
      ['formData.status']: status.value
    })
  },

  async submitForm() {
    const { formData, isEdit } = this.data
    if (!formData.name) {
      wx.showToast({ title: '请输入公司名称', icon: 'none' })
      return
    }

    const submitData = {
      name: formData.name,
      description: formData.description || '',
      status: formData.status || 'active'
    }

    this.setData({ loading: true })
    try {
      if (isEdit) {
        await companyApi.update(formData.id, submitData)
      } else {
        await companyApi.create(submitData)
      }
      wx.showToast({ title: isEdit ? '更新成功' : '创建成功', icon: 'success' })
      this.hideForm()
      this.setData({ current: 1, companies: [] })
      this.loadCompanies()
    } catch (error) {
      console.error('保存公司失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async deleteCompany(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该公司吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await companyApi.delete(id)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.setData({ current: 1, companies: [] })
            this.loadCompanies()
          } catch (error) {
            console.error('删除公司失败:', error)
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  goBack() {
    wx.navigateBack()
  },

  getStatusName(status) {
    return status === 'active' ? '正常' : '禁用'
  }
})
