import { shareApi, activityApi } from '../../../utils/request'

Page({
  data: {
    loading: false,
    // Activities for generating share
    activities: [],
    allActivities: [],
    selectedActivity: null,
    showActivityPicker: false,
    // Generated share
    shareData: null,
    showShareModal: false,
    showQRCodeModal: false,
    // Share records
    records: [],
    recordsLoading: false,
    recordsCurrent: 1,
    recordsTotal: 0,
    recordsHasMore: true,
    // Active tab
    activeTab: 'generate'
  },

  goBack() {
    wx.navigateBack()
  },

  onLoad() {
    this.loadActivities()
    this.loadShareRecords()
  },

  switchTab(e) {
    const tab = e.currentTarget.dataset.tab
    this.setData({ activeTab: tab })
  },

  async loadActivities() {
    try {
      const res = await activityApi.list(1, 100)
      if (res.success && res.data) {
        const list = (res.data.activities || []).map(item => ({
          id: item.id,
          name: item.name || item.pageTitle || '未命名活动',
          pageTitle: item.pageTitle || item.name || ''
        }))
        this.setData({
          activities: list,
          allActivities: list
        })
      }
    } catch (error) {
      console.error('加载活动列表失败:', error)
    }
  },

  async loadShareRecords(loadMore = false) {
    const { recordsCurrent } = this.data
    this.setData({ recordsLoading: true })
    try {
      const res = await shareApi.records(recordsCurrent, 20)
      if (res.success && res.data) {
        const list = (res.data.records || []).map(r => ({
          id: r.id,
          activityId: r.activityId,
          activityName: r.activityName || '未知活动',
          shareType: r.shareType || 'link',
          shareUrl: r.shareUrl || '',
          visitCount: r.visitCount || 0,
          createTime: this.formatDate(r.createTime)
        }))
        this.setData({
          records: loadMore ? [...this.data.records, ...list] : list,
          recordsTotal: res.data.pagination?.totalCount || 0,
          recordsHasMore: list.length >= 20
        })
      }
    } catch (error) {
      console.error('加载分享记录失败:', error)
    } finally {
      this.setData({ recordsLoading: false })
    }
  },

  formatDate(dateStr) {
    if (!dateStr) return ''
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    const hours = String(d.getHours()).padStart(2, '0')
    const minutes = String(d.getMinutes()).padStart(2, '0')
    return `${year}/${month}/${day} ${hours}:${minutes}`
  },

  onReachBottom() {
    if (this.data.recordsHasMore && !this.data.recordsLoading) {
      this.setData({ recordsCurrent: this.data.recordsCurrent + 1 })
      this.loadShareRecords(true)
    }
  },

  // ==================== 生成分享 ====================
  showActivitySelector() {
    this.setData({ showActivityPicker: true })
  },

  hideActivitySelector() {
    this.setData({ showActivityPicker: false })
  },

  onActivitySelect(e) {
    const index = e.currentTarget.dataset.index
    const activity = this.data.activities[index]
    this.setData({
      selectedActivity: activity,
      showActivityPicker: false,
      shareData: null,
      showShareModal: false,
      showQRCodeModal: false
    })
  },

  async generateLinkShare() {
    const { selectedActivity } = this.data
    if (!selectedActivity) {
      wx.showToast({ title: '请先选择活动', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    try {
      const res = await shareApi.generate({
        activityId: selectedActivity.id,
        shareType: 'link'
      })
      if (res.success && res.data) {
        this.setData({
          shareData: {
            ...res.data,
            activityTitle: selectedActivity.name,
            activityId: selectedActivity.id
          },
          showShareModal: true,
          showQRCodeModal: false
        })
      }
    } catch (error) {
      console.error('生成分享链接失败:', error)
      wx.showToast({ title: '生成失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  async generateQRCode() {
    const { selectedActivity } = this.data
    if (!selectedActivity) {
      wx.showToast({ title: '请先选择活动', icon: 'none' })
      return
    }
    this.setData({ loading: true })
    try {
      const res = await shareApi.generate({
        activityId: selectedActivity.id,
        shareType: 'qrcode'
      })
      if (res.success && res.data) {
        this.setData({
          shareData: {
            ...res.data,
            activityTitle: selectedActivity.name,
            activityId: selectedActivity.id
          },
          showQRCodeModal: true,
          showShareModal: false
        })
      }
    } catch (error) {
      console.error('生成二维码失败:', error)
      wx.showToast({ title: '生成失败', icon: 'none' })
    } finally {
      this.setData({ loading: false })
    }
  },

  hideShareModal() {
    this.setData({ showShareModal: false })
  },

  hideQRCodeModal() {
    this.setData({ showQRCodeModal: false })
  },

  async copyShareLink() {
    const { shareData } = this.data
    if (!shareData || !shareData.shareUrl) {
      wx.showToast({ title: '分享链接不存在', icon: 'none' })
      return
    }
    wx.setClipboardData({
      data: shareData.shareUrl,
      success: () => {
        wx.showToast({ title: '已复制', icon: 'success' })
      },
      fail: () => {
        wx.showToast({ title: '复制失败', icon: 'none' })
      }
    })
  },

  async saveQRCodeImage() {
    const { shareData } = this.data
    if (!shareData || !shareData.qrCodeImage) {
      wx.showToast({ title: '二维码不存在', icon: 'none' })
      return
    }
    try {
      const filePath = `${wx.env.USER_DATA_PATH}/share_qrcode_${shareData.id || Date.now()}.png`
      const fs = wx.getFileSystemManager()
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
              wx.showToast({ title: '已保存到相册', icon: 'success' })
            },
            fail: () => {
              wx.previewImage({ urls: [filePath], fail: () => {
                wx.showToast({ title: '保存失败', icon: 'none' })
              }})
            }
          })
        },
        fail: () => {
          wx.showToast({ title: '保存失败', icon: 'none' })
        }
      })
    } catch (error) {
      console.error('保存二维码失败:', error)
      wx.showToast({ title: '保存失败', icon: 'none' })
    }
  },

  previewQRCode() {
    const { shareData } = this.data
    if (!shareData || !shareData.qrCodeImage) return
    const filePath = `${wx.env.USER_DATA_PATH}/preview_qrcode.png`
    const fs = wx.getFileSystemManager()
    const base64Data = shareData.qrCodeImage.replace(/^data:image\/\w+;base64,/, '')
    const buffer = wx.base64ToArrayBuffer(base64Data)
    fs.writeFile({
      filePath,
      data: buffer,
      encoding: 'binary',
      success: () => {
        wx.previewImage({ urls: [filePath] })
      }
    })
  }
})
