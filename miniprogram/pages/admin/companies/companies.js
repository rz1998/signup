import { companyApi, userApi } from '../../../utils/request'

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
    ],
    // User picker
    showUserPicker: false,
    userList: [],
    filteredUserList: [],
    selectedAdminId: '',
    selectedAdmin: null,
    userSearchKey: '',
    // Quick add user
    showQuickAddUser: false,
    quickAddData: {}
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
      statusLabel: '正常',
      selectedAdminId: '',
      selectedAdmin: null,
      showQuickAddUser: false,
      quickAddData: {}
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
      statusLabel: company.status === 'inactive' ? '禁用' : '正常',
      selectedAdminId: company.adminUserId || '',
      selectedAdmin: company.adminUserId ? { id: company.adminUserId, name: company.adminUserName, phone: company.adminUserPhone } : null,
      showQuickAddUser: false,
      quickAddData: {}
    })
  },

  hideForm() {
    this.setData({ showForm: false, formData: {}, selectedAdminId: '', selectedAdmin: null })
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
    const { formData, isEdit, selectedAdmin } = this.data
    if (!formData.name) {
      wx.showToast({ title: '请输入公司名称', icon: 'none' })
      return
    }

    const submitData = {
      name: formData.name,
      description: formData.description || '',
      adminUserId: selectedAdmin?.id || '',
      adminUserName: selectedAdmin?.name || '',
      adminUserPhone: selectedAdmin?.phone || '',
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
  },

  // 用户选择相关
  async loadUsers() {
    try {
      // 公司管理员只能由 company_admin 角色的用户担任
      const res = await userApi.list(1, 100, { role: 'company_admin' })
      if (res.success && res.data) {
        const list = res.data.users || []
        this.setData({
          userList: list,
          filteredUserList: list
        })
      }
    } catch (error) {
      console.error('加载用户列表失败:', error)
    }
  },

  showUserSelector() {
    this.loadUsers()
    this.setData({
      showUserPicker: true,
      userSearchKey: '',
      filteredUserList: this.data.userList
    })
  },

  hideUserPicker() {
    this.setData({ showUserPicker: false, showQuickAddUser: false, quickAddData: {} })
  },

  stopPropagation() {
    // 阻止事件冒泡
  },

  onUserSearch(e) {
    const key = e.detail.value.toLowerCase()
    const { userList } = this.data
    const filtered = userList.filter(user =>
      !key || (user.name && user.name.toLowerCase().includes(key)) ||
        (user.phone && user.phone.includes(key))
    )
    this.setData({
      userSearchKey: key,
      filteredUserList: filtered
    })
  },

  selectAdmin(e) {
    const user = e.currentTarget.dataset.user
    this.setData({
      selectedAdminId: user.id,
      selectedAdmin: user,
      showUserPicker: false
    })
  },

  clearSelectedAdmin() {
    this.setData({
      selectedAdminId: '',
      selectedAdmin: null
    })
  },

  showQuickAddUser() {
    this.setData({
      showQuickAddUser: true,
      quickAddData: {}
    })
  },

  cancelQuickAddUser() {
    this.setData({ showQuickAddUser: false, quickAddData: {} })
  },

  onQuickAddInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`quickAddData.${field}`]: value })
  },

  async confirmQuickAddUser() {
    const { quickAddData } = this.data
    if (!quickAddData.name) {
      wx.showToast({ title: '请输入姓名', icon: 'none' })
      return
    }
    if (!quickAddData.phone) {
      wx.showToast({ title: '请输入电话', icon: 'none' })
      return
    }
    if (!quickAddData.password) {
      wx.showToast({ title: '请输入密码', icon: 'none' })
      return
    }

    try {
      const res = await userApi.create({
        name: quickAddData.name,
        phone: quickAddData.phone,
        password: quickAddData.password,
        role: 'company_admin'
      })
      if (res.success && res.data) {
        wx.showToast({ title: '创建成功', icon: 'success' })
        const newUser = res.data
        this.setData({
          showQuickAddUser: false,
          quickAddData: {},
          selectedAdminId: newUser.id,
          selectedAdmin: newUser
        })
        this.loadUsers()
        this.setData({ showUserPicker: false })
      } else {
        wx.showToast({ title: res.message || '创建失败', icon: 'none' })
      }
    } catch (error) {
      console.error('创建用户失败:', error)
      wx.showToast({ title: '创建失败', icon: 'none' })
    }
  }
})
