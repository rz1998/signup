import { branchApi, companyApi } from '../../../utils/request'

Page({
  data: {
    branches: [],
    loading: false,
    current: 1,
    total: 0,
    hasMore: true,
    keyword: '',
    // Filter
    companies: [],
    filterCompanyId: '',
    filterCompanyIndex: 0,
    filterCompanyLabel: '全部公司',
    // Form
    showForm: false,
    formData: {},
    isEdit: false,
    formCompanies: [],
    formCompanyIndex: 0,
    formCompanyLabel: '请选择所属公司',
    statusIndex: 0,
    statusLabel: '正常',
    statuses: [
      { label: '正常', value: 'active' },
      { label: '禁用', value: 'inactive' }
    ]
  },

  onLoad() {
    const app = getApp()
    if (!app.isAdmin() && !app.isMgr()) {
      wx.showToast({ title: '无权限访问', icon: 'none' })
      wx.navigateBack()
      return
    }
    this.loadCompanies()
    this.loadBranches()
  },

  async loadCompanies() {
    try {
      const res = await companyApi.list(1, 100)
      if (res.success && res.data) {
        const list = res.data.companies || []
        this.setData({
          companies: list,
          formCompanies: list
        })
      }
    } catch (error) {
      console.error('加载公司列表失败:', error)
    }
  },

  async loadBranches(loadMore = false) {
    const { current, branches, filterCompanyId, keyword } = this.data
    this.setData({ loading: true })
    try {
      const res = await branchApi.list(current, 20, filterCompanyId, keyword)
      if (res.success && res.data) {
        const list = res.data.branches || []
        this.setData({
          branches: loadMore ? [...branches, ...list] : list,
          total: res.data.pagination?.totalCount || 0,
          hasMore: list.length >= 20
        })
      }
    } catch (error) {
      console.error('加载分支机构列表失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ current: this.data.current + 1 })
      this.loadBranches(true)
    }
  },

  onSearch(e) {
    const keyword = e.detail.value
    this.setData({ keyword, current: 1, branches: [] })
    this.loadBranches()
  },

  onFilterCompanyChange(e) {
    const index = parseInt(e.detail.value, 10)
    const companies = [{ id: '', name: '全部公司' }, ...this.data.companies]
    const company = companies[index]
    const companyId = company ? company.id : ''
    this.setData({
      filterCompanyIndex: index,
      filterCompanyId: companyId,
      filterCompanyLabel: company ? company.name : '全部公司',
      current: 1,
      branches: []
    })
    this.loadBranches()
  },

  showAddForm() {
    this.setData({
      showForm: true,
      formData: {},
      isEdit: false,
      formCompanies: this.data.companies,
      formCompanyIndex: 0,
      formCompanyLabel: '请选择所属公司',
      statusIndex: 0,
      statusLabel: '正常'
    })
  },

  showEditForm(e) {
    const branch = e.currentTarget.dataset.branch
    const companies = this.data.companies

    let formCompanyIndex = 0
    let formCompanyLabel = '请选择所属公司'
    if (branch.companyId) {
      for (let i = 0; i < companies.length; i++) {
        if (companies[i].id === branch.companyId) {
          formCompanyIndex = i
          formCompanyLabel = companies[i].name
          break
        }
      }
    }

    const statusIndex = branch.status === 'inactive' ? 1 : 0
    this.setData({
      showForm: true,
      formData: {
        id: branch.id,
        name: branch.name || '',
        leaderName: branch.leaderName || '',
        leaderPhone: branch.leaderPhone || '',
        description: branch.description || ''
      },
      isEdit: true,
      formCompanies: companies,
      formCompanyIndex,
      formCompanyLabel,
      statusIndex,
      statusLabel: branch.status === 'inactive' ? '禁用' : '正常'
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

  onFormCompanyChange(e) {
    const index = parseInt(e.detail.value, 10)
    const company = this.data.formCompanies[index]
    this.setData({
      formCompanyIndex: index,
      formCompanyLabel: company ? company.name : '请选择所属公司',
      ['formData.companyId']: company ? company.id : ''
    })
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
      wx.showToast({ title: '请输入分支机构名称', icon: 'none' })
      return
    }
    if (!formData.companyId) {
      wx.showToast({ title: '请选择所属公司', icon: 'none' })
      return
    }

    const submitData = {
      name: formData.name,
      companyId: formData.companyId,
      leaderName: formData.leaderName || '',
      leaderPhone: formData.leaderPhone || '',
      description: formData.description || '',
      status: formData.status || 'active'
    }

    this.setData({ loading: true })
    try {
      if (isEdit) {
        await branchApi.update(formData.id, submitData)
      } else {
        await branchApi.create(submitData)
      }
      wx.showToast({ title: isEdit ? '更新成功' : '创建成功', icon: 'success' })
      this.hideForm()
      this.setData({ current: 1, branches: [] })
      this.loadBranches()
    } catch (error) {
      console.error('保存分支机构失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async deleteBranch(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该分支机构吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await branchApi.delete(id)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.setData({ current: 1, branches: [] })
            this.loadBranches()
          } catch (error) {
            console.error('删除分支机构失败:', error)
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  goBack() {
    wx.navigateBack()
  }
})
