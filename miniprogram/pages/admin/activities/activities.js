import { activityApi, shareApi, branchApi, companyApi } from '../../../utils/request'

Page({
  // 日期格式化
  formatDate(dateStr) {
    if (!dateStr) return ''
    // 支持 ISO 格式和普通格式
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    const hours = String(d.getHours()).padStart(2, '0')
    const minutes = String(d.getMinutes()).padStart(2, '0')
    const seconds = String(d.getSeconds()).padStart(2, '0')
    return `${year}年${month}月${day}日 ${hours}:${minutes}:${seconds}`
  },

  formatDateShort(dateStr) {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}年${month}月${day}日`
  },

  data: {
    canManageActivities: false,
    canDeleteActivities: false,
    activities: [],
    loading: false,
    current: 1,
    total: 0,
    hasMore: true,
    // Branch filter
    companies: [],
    branches: [],
    filterCompanyId: '',
    filterBranchId: '',
    filterCompanyIndex: 0,
    filterBranchIndex: 0,
    filterCompanyLabel: '全部公司',
    filterBranchLabel: '全部分支机构',
    // Form states
    showForm: false,
    formData: {},
    isEdit: false,
    // Date/time picker states
    startDate: '',
    startTime: '',
    endDate: '',
    endTime: '',
    // Settings states
    showSettings: false,
    settingsActivity: {},
    settingsData: {},
    // Share states
    shareActivity: null
  },

  onLoad() {
    const app = getApp()
    const isMgr = app.isMgr()
    const isAdmin = app.isAdmin()
    this.setData({
      canManageActivities: isMgr,
      canDeleteActivities: isAdmin
    })
    this.loadCompanies()
    this.loadActivities()
  },

  async loadCompanies() {
    try {
      const res = await companyApi.list(1, 100)
      if (res.success && res.data) {
        const list = res.data.companies || []
        this.setData({ companies: list })
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
        this.setData({ branches: list })
      }
    } catch (error) {
      console.error('加载分支机构列表失败:', error)
    }
  },

  async loadActivities(loadMore = false) {
    const { current, activities, filterBranchId } = this.data
    const params = { page: current, pageSize: 10 }
    if (filterBranchId) params.branchId = filterBranchId

    this.setData({ loading: true })
    try {
      const res = await activityApi.list(current, 10, params)
      if (res.success && res.data) {
        // 直接使用API返回的字段，并用格式化函数处理日期
        const list = (res.data.activities || []).map(item => ({
          id: item.id,
          name: item.name || '',
          title: item.name || '',
          description: item.description || '',
          coverImage: item.coverImage || '',
          location: item.location || '',
          maxParticipants: item.maxParticipants || 0,
          currentParticipants: item.currentParticipants || 0,
          startDate: item.startDate || '',
          endDate: item.endDate || '',
          startTime: item.startDate ? this.formatDate(item.startDate) : '',
          endTime: item.endDate ? this.formatDate(item.endDate) : '',
          startDateShort: item.startDate ? this.formatDateShort(item.startDate) : '',
          endDateShort: item.endDate ? this.formatDateShort(item.endDate) : '',
          status: item.status || 'draft',
          pageTitle: item.pageTitle || '',
          pageSubtitle: item.pageSubtitle || ''
        }))
        
        this.setData({
          activities: loadMore ? [...activities, ...list] : list,
          total: res.data.pagination?.totalCount || 0,
          hasMore: list.length >= 10
        })
      }
    } catch (error) {
      console.error('加载活动列表失败:', error)
    } finally {
      this.setData({ loading: false })
    }
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.setData({
        current: this.data.current + 1
      })
      this.loadActivities(true)
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
      activities: []
    })
    this.loadBranches(companyId)
    this.loadActivities()
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
      activities: []
    })
    this.loadActivities()
  },

  showAddForm() {
    if (!this.data.canManageActivities) {
      wx.showToast({ title: '无权限', icon: 'none' })
      return
    }
    this.setData({
      showForm: true,
      formData: {},
      isEdit: false,
      startDate: '',
      startTime: '',
      endDate: '',
      endTime: ''
    })
  },

  showEditForm(e) {
    if (!this.data.canManageActivities) {
      wx.showToast({ title: '无权限', icon: 'none' })
      return
    }
    const activity = e.currentTarget.dataset.activity
    // Parse existing datetime strings (API returns startDate/endDate)
    let startDate = '', startTime = '', endDate = '', endTime = ''
    // API returns startDate/endDate as datetime strings like "2026-03-19T10:00:00Z"
    if (activity.startDate) {
      const startParts = activity.startDate.split('T')
      startDate = startParts[0] || ''
      startTime = startParts[1] ? startParts[1].substring(0, 5) : ''
    }
    if (activity.endDate) {
      const endParts = activity.endDate.split('T')
      endDate = endParts[0] || ''
      endTime = endParts[1] ? endParts[1].substring(0, 5) : ''
    }
    
    this.setData({
      showForm: true,
      formData: { 
        id: activity.id,
        name: activity.name || activity.title,
        description: activity.description || '',
        location: activity.location || '',
        maxParticipants: activity.maxParticipants || 100
      },
      isEdit: true,
      startDate,
      startTime,
      endDate,
      endTime
    })
  },

  hideForm() {
    this.setData({
      showForm: false,
      formData: {},
      startDate: '',
      startTime: '',
      endDate: '',
      endTime: ''
    })
  },

  showSettingsForm(e) {
    const activity = e.currentTarget.dataset.activity
    this.setData({
      showSettings: true,
      settingsActivity: activity,
      settingsData: {
        pageTitle: activity.pageTitle || '',
        pageSubtitle: activity.pageSubtitle || '',
        pageDescription: activity.pageDescription || '',
        registrationNotice: activity.registrationNotice || '',
        contactPhone: activity.contactPhone || '',
        contactEmail: activity.contactEmail || ''
      }
    })
  },

  hideSettingsForm() {
    this.setData({
      showSettings: false,
      settingsActivity: {},
      settingsData: {}
    })
  },

  onSettingsInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({
      [`settingsData.${field}`]: value
    })
  },

  async saveSettings() {
    const { settingsActivity, settingsData } = this.data
    
    this.setData({ loading: true })
    try {
      await activityApi.update(settingsActivity.id, {
        pageTitle: settingsData.pageTitle,
        pageSubtitle: settingsData.pageSubtitle,
        pageDescription: settingsData.pageDescription,
        registrationNotice: settingsData.registrationNotice,
        contactPhone: settingsData.contactPhone,
        contactEmail: settingsData.contactEmail
      })
      
      wx.showToast({
        title: '保存成功',
        icon: 'success'
      })
      
      this.hideSettingsForm()
      this.setData({
        current: 1,
        activities: []
      })
      this.loadActivities()
    } catch (error) {
      console.error('保存页面设置失败:', error)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  onFormInput(e) {
    const field = e.currentTarget.dataset.field
    const value = e.detail.value
    this.setData({
      [`formData.${field}`]: value
    })
  },

  onStartDateChange(e) {
    const startDate = e.detail.value
    this.setData({ startDate })
    this.updateDateTime('start')
  },

  onStartTimeChange(e) {
    const startTime = e.detail.value
    this.setData({ startTime })
    this.updateDateTime('start')
  },

  onEndDateChange(e) {
    const endDate = e.detail.value
    this.setData({ endDate })
    this.updateDateTime('end')
  },

  onEndTimeChange(e) {
    const endTime = e.detail.value
    this.setData({ endTime })
    this.updateDateTime('end')
  },

  updateDateTime(type) {
    const { startDate, startTime, endDate, endTime } = this.data
    const formData = { ...this.data.formData }
    
    if (type === 'start' && startDate) {
      formData.startTime = startTime ? `${startDate} ${startTime}` : `${startDate} 00:00`
    }
    if (type === 'end' && endDate) {
      formData.endTime = endTime ? `${endDate} ${endTime}` : `${endDate} 23:59`
    }
    
    this.setData({ formData })
  },

  async submitForm() {
    const { formData, isEdit, startDate, startTime, endDate, endTime } = this.data
    
    if (!formData.name) {
      wx.showToast({
        title: '请填写活动标题',
        icon: 'none'
      })
      return
    }

    if (!startDate || !endDate) {
      wx.showToast({
        title: '请选择开始和结束时间',
        icon: 'none'
      })
      return
    }

    // Convert to API format (ISO datetime string)
    const apiData = {
      name: formData.name,
      description: formData.description || '',
      location: formData.location || '',
      maxParticipants: formData.maxParticipants || 100,
      startDate: startTime ? `${startDate}T${startTime}:00` : `${startDate}T00:00:00`,
      endDate: endTime ? `${endDate}T${endTime}:00` : `${endDate}T23:59:59`,
      status: 'draft'
    }

    this.setData({ loading: true })
    try {
      if (isEdit && formData.id) {
        await activityApi.update(formData.id, apiData)
      } else {
        await activityApi.create(apiData)
      }
      
      wx.showToast({
        title: isEdit ? '更新成功' : '创建成功',
        icon: 'success'
      })
      
      this.hideForm()
      this.setData({
        current: 1,
        activities: []
      })
      this.loadActivities()
    } catch (error) {
      console.error('保存活动失败:', error)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async deleteActivity(e) {
    if (!this.data.canDeleteActivities) {
      wx.showToast({ title: '无权限', icon: 'none' })
      return
    }
    const id = e.currentTarget.dataset.id
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除该活动吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await activityApi.delete(id)
            wx.showToast({
              title: '删除成功',
              icon: 'success'
            })
            this.setData({
              current: 1,
              activities: []
            })
            this.loadActivities()
          } catch (error) {
            console.error('删除活动失败:', error)
          }
        }
      }
    })
  },

  goBack() {
    wx.navigateBack()
  },

  goToSettings(e) {
    const activity = e.currentTarget.dataset.activity
    wx.navigateTo({
      url: `/pages/admin/settings/settings?activityId=${activity.id}&activityTitle=${encodeURIComponent(activity.title)}`
    })
  },

  goToConfig(e) {
    const activity = e.currentTarget.dataset.activity
    wx.navigateTo({
      url: `/pages/admin/activity-config/activity-config?activityId=${activity.id}&activityTitle=${encodeURIComponent(activity.pageTitle || activity.name)}`
    })
  },

  goToDetail(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/activity/activity?id=${id}`
    })
  },

  // ==================== 分享功能 ====================
  // 阻止事件冒泡
  preventBubble() {},

  // 存储要分享的活动
  onSelectShare(e) {
    const activity = e.currentTarget.dataset.activity
    this.setData({ shareActivity: activity })
  },

  // 分享小程序卡片
  onShareAppMessage(e) {
    const activity = this.data.shareActivity
    if (!activity || !activity.id) {
      return {}
    }
    return {
      title: activity.pageTitle || activity.name || '活动报名',
      path: `/pages/activity/activity?id=${activity.id}`,
      imageUrl: activity.coverImage || ''
    }
  },

  async showShareModal(e) {
    const activity = e.currentTarget.dataset.activity
    if (!activity || !activity.id) {
      wx.showToast({
        title: '活动信息不完整',
        icon: 'none'
      })
      return
    }

    this.setData({ loading: true })
    try {
      const res = await shareApi.generate({
        activityId: activity.id,
        shareType: 'link'
      })
      if (res.success && res.data) {
        this.setData({
          showShare: true,
          shareData: {
            ...res.data,
            activityId: activity.id,
            activityTitle: activity.title
          }
        })
      }
    } catch (error) {
      console.error('生成分享链接失败:', error)
      wx.showToast({
        title: '生成分享链接失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  async showQRCodeModal(e) {
    const activity = e.currentTarget.dataset.activity
    if (!activity || !activity.id) {
      wx.showToast({
        title: '活动信息不完整',
        icon: 'none'
      })
      return
    }

    this.setData({ loading: true })
    try {
      const res = await shareApi.generate({
        activityId: activity.id,
        shareType: 'qrcode'
      })
      if (res.success && res.data) {
        this.setData({
          showQRCode: true,
          shareData: {
            ...res.data,
            activityId: activity.id,
            activityTitle: activity.title
          }
        })
      }
    } catch (error) {
      console.error('生成二维码失败:', error)
      wx.showToast({
        title: '生成二维码失败',
        icon: 'none'
      })
    } finally {
      this.setData({ loading: false })
    }
  },

  hideShareModal() {
    this.setData({
      showShare: false,
      shareData: {}
    })
  },

  hideQRCodeModal() {
    this.setData({
      showQRCode: false,
      shareData: {}
    })
  },

  async copyShareLink() {
    const { shareData } = this.data
    if (!shareData.shareUrl) {
      wx.showToast({
        title: '分享链接不存在',
        icon: 'none'
      })
      return
    }

    wx.setClipboardData({
      data: shareData.shareUrl,
      success: () => {
        wx.showToast({
          title: '链接已复制',
          icon: 'success'
        })
      },
      fail: () => {
        wx.showToast({
          title: '复制失败',
          icon: 'none'
        })
      }
    })
  },

  async saveQRCode() {
    const { shareData } = this.data
    if (!shareData.qrCodeImage) {
      wx.showToast({
        title: '二维码不存在',
        icon: 'none'
      })
      return
    }

    // Convert base64 to temp file and save
    try {
      const filePath = `${wx.env.USER_DATA_PATH}/qrcode_${shareData.id}.png`
      const fs = wx.getFileSystemManager()

      // Remove data URL prefix if present
      const base64Data = shareData.qrCodeImage.replace(/^data:image\/\w+;base64,/, '')
      const buffer = wx.base64ToArrayBuffer(base64Data)

      fs.writeFile({
        filePath,
        data: buffer,
        encoding: 'binary',
        success: () => {
          wx.saveImageToPhotosAlbum({
            filePath,
            success: () => {
              wx.showToast({
                title: '二维码已保存',
                icon: 'success'
              })
            },
            fail: (err) => {
              console.error('保存失败:', err)
              // Fallback: show preview for manual save
              wx.previewImage({
                urls: [filePath],
                fail: () => {
                  wx.showToast({
                    title: '保存失败',
                    icon: 'none'
                  })
                }
              })
            }
          })
        },
        fail: (err) => {
          console.error('写入文件失败:', err)
          wx.showToast({
            title: '保存失败',
            icon: 'none'
          })
        }
      })
    } catch (error) {
      console.error('保存二维码失败:', error)
      wx.showToast({
        title: '保存失败',
        icon: 'none'
      })
    }
  }
})
