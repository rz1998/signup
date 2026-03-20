import { userApi, companyApi, branchApi } from '../../../utils/request'

Page({
  data: {
    users: [],
    loading: false,
    current: 1,
    total: 0,
    hasMore: true,
    // Filter states
    filterCompanyId: '',
    filterBranchId: '',
    filterCompanyIndex: 0,
    filterBranchIndex: 0,
    companies: [],
    branches: [],
    managers: [],
    showFilterCompany: false,
    showFilterBranch: false,
    filterCompanyLabel: '全部公司',
    filterBranchLabel: '全部分支机构',
    // Form states
    showForm: false,
    formData: {},
    isEdit: false,
    roles: [
      { label: '管理员', value: 'admin' },
      { label: '公司管理员', value: 'company_mgr' },
      { label: '分支机构管理员', value: 'branch_mgr' },
      { label: '营销管理', value: 'mgr' },
      { label: '营销人员', value: 'sales' }
    ],
    statuses: [
      { label: '正常', value: 'active' },
      { label: '禁用', value: 'inactive' }
    ],
    roleIndex: 0,
    roleLabel: '营销人员',
    statusIndex: 0,
    statusLabel: '正常',
    // Form pickers
    formCompanyIndex: 0,
    formBranchIndex: 0,
    formManagerIndex: 0,
    formCompanyLabel: '请选择公司',
    formBranchLabel: '请选择分支机构',
    formManagerLabel: '无（仅普通营销人员）',
    formCompanies: [],
    formBranches: [],
    formManagers: []
  },

  onLoad() {
    const app = getApp()
    if (!app.isAdmin() && !app.isMgr()) {
      wx.showToast({ title: '无权限访问', icon: 'none' })
      wx.navigateBack()
      return
    }
    this.loadCompanies()
    this.loadUsers()
  },

  async loadCompanies() {
    try {
      const res = await companyApi.list(1, 100)
      if (res.success && res.data) {
        const list = res.data.companies || []
        this.setData({
          companies: list,
          formCompanies: [{ id: '', name: '无（系统级）' }, ...list]
        })
      }
    } catch (error) {
      console.error('加载公司列表失败:', error)
    }
  },

  async loadBranches(companyId) {
    try {
      const res = await branchApi.list(1, 100, companyId || '')
      if (res.success && res.data) {
        const list = res.data.branches || []
        this.setData({
          branches: list,
          formBranches: [{ id: '', name: '无' }, ...list]
        })
      }
    } catch (error) {
      console.error('加载分支机构列表失败:', error)
    }
  },

  async loadUsers(loadMore = false) {
    const { current, users, filterCompanyId, filterBranchId } = this.data
    const params = { page: current, pageSize: 20 }
    if (filterCompanyId) params.companyId = filterCompanyId
    if (filterBranchId) params.branchId = filterBranchId

    this.setData({ loading: true })
    try {
      const res = await userApi.list(current, 20, params)
      if (res.success && res.data) {
        const list = res.data.users || []
        this.setData({
          users: loadMore ? [...users, ...list] : list,
          total: res.data.pagination?.totalCount || 0,
          hasMore: list.length >= 20
        })
      }
    } catch (error) {
      console.error('加载用户列表失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({ current: this.data.current + 1 })
      this.loadUsers(true)
    }
  },

  // ==================== 筛选器 ====================
  onFilterCompanyChange(e) {
    const index = parseInt(e.detail.value, 10)
    const companies = [{ id: '', name: '全部公司' }, ...this.data.companies]
    const company = companies[index]
    const companyId = company ? company.id : ''
    this.setData({
      filterCompanyIndex: index,
      filterCompanyId: companyId,
      filterCompanyLabel: company ? company.name : '全部公司',
      filterBranchId: '',
      filterBranchIndex: 0,
      filterBranchLabel: '全部分支机构',
      current: 1,
      users: []
    })
    this.loadBranches(companyId)
    this.loadUsers()
  },

  onFilterBranchChange(e) {
    const index = parseInt(e.detail.value, 10)
    const branches = [{ id: '', name: '全部分支机构' }, ...this.data.branches]
    const branch = branches[index]
    const branchId = branch ? branch.id : ''
    this.setData({
      filterBranchIndex: index,
      filterBranchId: branchId,
      filterBranchLabel: branch ? branch.name : '全部分支机构',
      current: 1,
      users: []
    })
    this.loadUsers()
  },

  // ==================== 表单 ====================
  showAddForm() {
    this.setData({
      showForm: true,
      formData: {},
      isEdit: false,
      roleIndex: 4,
      roleLabel: '营销人员',
      statusIndex: 0,
      statusLabel: '正常',
      formCompanyIndex: 0,
      formBranchIndex: 0,
      formManagerIndex: 0,
      formCompanyLabel: '请选择公司',
      formBranchLabel: '请选择分支机构',
      formManagerLabel: '无（仅普通营销人员）',
      formCompanies: [{ id: '', name: '无（系统级）' }, ...this.data.companies],
      formBranches: [{ id: '', name: '无' }],
      formManagers: [{ id: '', name: '无' }],
      formBranchDisabled: false
    })
  },

  async showEditForm(e) {
    const user = e.currentTarget.dataset.user
    const roles = this.data.roles
    const statuses = this.data.statuses

    let roleIndex = 4
    let roleLabel = '营销人员'
    for (let i = 0; i < roles.length; i++) {
      if (roles[i].value === user.role) {
        roleIndex = i
        roleLabel = roles[i].label
        break
      }
    }

    let statusIndex = 0
    let statusLabel = '正常'
    for (let i = 0; i < statuses.length; i++) {
      if (statuses[i].value === user.status) {
        statusIndex = i
        statusLabel = statuses[i].label
        break
      }
    }

    // Load branches for the user's company
    await this.loadBranches(user.companyId || '')

    // Load managers for the user's branch
    let managers = [{ id: '', name: '无' }]
    if (user.branchId) {
      try {
        const res = await userApi.list(1, 100, { branchId: user.branchId, role: 'mgr' })
        if (res.success && res.data) {
          managers = [{ id: '', name: '无' }, ...(res.data.users || [])]
        }
      } catch (err) {
        console.error('加载管理人员列表失败:', err)
      }
    }

    const formCompanies = [{ id: '', name: '无（系统级）' }, ...this.data.companies]
    const formBranches = [{ id: '', name: '无' }, ...this.data.branches]
    const formManagers = managers

    // Find company index
    let formCompanyIndex = 0
    let formCompanyLabel = '请选择公司'
    if (user.companyId) {
      for (let i = 0; i < formCompanies.length; i++) {
        if (formCompanies[i].id === user.companyId) {
          formCompanyIndex = i
          formCompanyLabel = formCompanies[i].name
          break
        }
      }
    }

    // Find branch index
    let formBranchIndex = 0
    let formBranchLabel = '请选择分支机构'
    if (user.branchId) {
      for (let i = 0; i < formBranches.length; i++) {
        if (formBranches[i].id === user.branchId) {
          formBranchIndex = i
          formBranchLabel = formBranches[i].name
          break
        }
      }
    }

    // Find manager index
    let formManagerIndex = 0
    let formManagerLabel = '无（仅普通营销人员）'
    if (user.managerId) {
      for (let i = 0; i < formManagers.length; i++) {
        if (formManagers[i].id === user.managerId) {
          formManagerIndex = i
          formManagerLabel = formManagers[i].name
          break
        }
      }
    }

    this.setData({
      showForm: true,
      formData: {
        id: user.id,
        username: user.username,
        phone: user.phone || '',
        email: user.email || '',
        fullName: user.fullName || '',
        role: user.role || 'sales',
        status: user.status || 'active',
        companyId: user.companyId || '',
        branchId: user.branchId || '',
        managerId: user.managerId || ''
      },
      isEdit: true,
      roleIndex,
      roleLabel,
      statusIndex,
      statusLabel,
      formCompanyIndex,
      formCompanyLabel,
      formBranchIndex,
      formBranchLabel,
      formManagerIndex,
      formManagerLabel,
      formCompanies,
      formBranches,
      formManagers
    })
  },

  hideForm() {
    this.setData({
      showForm: false,
      formData: {}
    })
  },

  onInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({ [`formData.${field}`]: value })
  },

  onRoleChange(e) {
    const index = parseInt(e.detail.value, 10)
    const role = this.data.roles[index]
    this.setData({
      roleIndex: index,
      roleLabel: role.label,
      ['formData.role']: role.value
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

  async onFormCompanyChange(e) {
    const index = parseInt(e.detail.value, 10)
    const companies = this.data.formCompanies
    const company = companies[index]
    const companyId = company ? company.id : ''

    this.setData({
      formCompanyIndex: index,
      formCompanyLabel: company ? company.name : '请选择公司',
      ['formData.companyId']: companyId,
      formBranchIndex: 0,
      formBranchLabel: '请选择分支机构',
      ['formData.branchId']: '',
      formManagerIndex: 0,
      formManagerLabel: '无（仅普通营销人员）',
      ['formData.managerId']: '',
      formManagers: [{ id: '', name: '无' }]
    })

    // Reload branches for selected company
    await this.loadBranches(companyId)
    this.setData({
      formBranches: [{ id: '', name: '无' }, ...this.data.branches]
    })
  },

  async onFormBranchChange(e) {
    const index = parseInt(e.detail.value, 10)
    const branches = this.data.formBranches
    const branch = branches[index]
    const branchId = branch ? branch.id : ''

    this.setData({
      formBranchIndex: index,
      formBranchLabel: branch ? branch.name : '请选择分支机构',
      ['formData.branchId']: branchId,
      formManagerIndex: 0,
      formManagerLabel: '无（仅普通营销人员）',
      ['formData.managerId']: ''
    })

    // Reload managers for selected branch
    if (branchId) {
      try {
        const res = await userApi.list(1, 100, { branchId, role: 'mgr' })
        if (res.success && res.data) {
          const managers = [{ id: '', name: '无' }, ...(res.data.users || [])]
          this.setData({ formManagers: managers })
        }
      } catch (err) {
        console.error('加载管理人员列表失败:', err)
      }
    } else {
      this.setData({ formManagers: [{ id: '', name: '无' }] })
    }
  },

  onFormManagerChange(e) {
    const index = parseInt(e.detail.value, 10)
    const managers = this.data.formManagers
    const manager = managers[index]
    this.setData({
      formManagerIndex: index,
      formManagerLabel: manager ? manager.name : '无（仅普通营销人员）',
      ['formData.managerId']: manager ? manager.id : ''
    })
  },

  async submitForm() {
    const { formData, isEdit } = this.data

    if (!formData.username && !isEdit) {
      wx.showToast({ title: '请输入用户名', icon: 'none' })
      return
    }

    if (!formData.password && !isEdit) {
      wx.showToast({ title: '请输入密码', icon: 'none' })
      return
    }

    // Clean up empty strings for optional fields
    const submitData = { ...formData }
    if (!submitData.companyId) delete submitData.companyId
    if (!submitData.branchId) delete submitData.branchId
    if (!submitData.managerId) delete submitData.managerId

    this.setData({ loading: true })
    try {
      if (isEdit) {
        await userApi.updateUser(submitData.id, submitData)
      } else {
        await userApi.create(submitData)
      }
      wx.showToast({ title: isEdit ? '更新成功' : '创建成功', icon: 'success' })
      this.hideForm()
      this.setData({ current: 1, users: [] })
      this.loadUsers()
    } catch (error) {
      console.error('保存用户失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async deleteUser(e) {
    const id = e.currentTarget.dataset.id
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该用户吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await userApi.delete(id)
            wx.showToast({ title: '删除成功', icon: 'success' })
            this.setData({ current: 1, users: [] })
            this.loadUsers()
          } catch (error) {
            console.error('删除用户失败:', error)
            wx.showToast({ title: '删除失败', icon: 'none' })
          }
        }
      }
    })
  },

  goBack() {
    wx.navigateBack()
  },

  getRoleName(role) {
    const map = {
      admin: '管理员',
      company_mgr: '公司管理员',
      branch_mgr: '分支机构管理员',
      mgr: '营销管理',
      sales: '营销人员'
    }
    return map[role] || role
  }
})
