const app = getApp()

const request = {
  getAuthHeader() {
    const token = wx.getStorageSync('token')
    if (token) {
      return {
        'Authorization': `Bearer ${token}`
      }
    }
    return {}
  },

  async request(options) {
    const { url, method = 'GET', data = {}, header = {} } = options
    const baseUrl = app.globalData.apiBaseUrl || 'http://192.168.8.230:8082/api/v1'
    const fullUrl = `${baseUrl}${url}`

    return new Promise((resolve, reject) => {
      wx.request({
        url: fullUrl,
        method,
        data,
        header: {
          'Content-Type': 'application/json',
          ...this.getAuthHeader(),
          ...header
        },
        success: (response) => {
          if (response.statusCode === 200) {
            resolve(response.data)
          } else if (response.statusCode === 401) {
            app.globalData.token = ''
            wx.removeStorageSync('token')
            wx.removeStorageSync('userInfo')
            wx.redirectTo({
              url: '/pages/login/login'
            })
            reject(new Error('未授权，请重新登录'))
          } else {
            reject(new Error(response.data?.message || response.data?.msg || '请求失败'))
          }
        },
        fail: (error) => {
          console.error('Request fail:', error)
          wx.showToast({
            title: error.errMsg || '网络错误',
            icon: 'none'
          })
          reject(error)
        }
      })
    })
  },

  get(url, data) {
    return this.request({ url, method: 'GET', data })
  },

  post(url, data) {
    return this.request({ url, method: 'POST', data })
  },

  put(url, data) {
    return this.request({ url, method: 'PUT', data })
  },

  delete(url, data) {
    return this.request({ url, method: 'DELETE', data })
  }
}

// 活动相关 API
export const activityApi = {
  list: (page = 1, pageSize = 10, params = {}) => 
    request.get('/activities', { page, pageSize, ...params }),
  get: (id) => 
    request.get(`/activities/${id}`),
  create: (data) => 
    request.post('/activities', data),
  update: (id, data) => 
    request.put(`/activities/${id}`, data),
  delete: (id) => 
    request.delete(`/activities/${id}`)
}

// 表单字段相关 API
export const formFieldApi = {
  list: (activityId) => 
    request.get(`/activities/${activityId}/fields`),
  create: (activityId, data) => 
    request.post(`/activities/${activityId}/fields`, data),
  update: (fieldId, data) => 
    request.put(`/fields/${fieldId}`, data),
  delete: (fieldId) => 
    request.delete(`/fields/${fieldId}`),
  reorder: (activityId, fieldIds) => 
    request.put(`/activities/${activityId}/fields/reorder`, { fieldIds })
}

// 报名相关 API
export const registrationApi = {
  create: (data) => 
    request.post('/registrations', data),
  // 报名列表 - sales角色只看自己的，admin/mgr看全部
  // 后端已根据JWT中的role自动过滤
  myList: (page = 1, pageSize = 10) => 
    request.get('/registrations', { page, pageSize }),
  cancel: (id) => 
    request.delete(`/registrations/${id}`),
  get: (id) => 
    request.get(`/registrations/${id}`),
  // 游客通过手机号查询自己的报名
  getByPhone: (phone) => 
    request.get('/visitor/registrations', { phone }),
  // 游客修改报名
  updateByVisitor: (id, data) =>
    request.put(`/visitor/registrations/${id}`, data)
}

// 用户相关 API
export const userApi = {
  login: (username, password) => 
    request.post('/auth/login', { username, password }),
  getInfo: () => 
    request.get('/users/me'),
  update: (data) => 
    request.put('/users/me', data),
  list: (page = 1, pageSize = 10, params = {}) => 
    request.get('/users', { page, pageSize, ...params }),
  create: (data) =>
    request.post('/users', data),
  updateUser: (id, data) =>
    request.put(`/users/${id}`, data),
  delete: (id) => 
    request.delete(`/users/${id}`)
}

// 管理员报名管理 API
export const adminRegistrationApi = {
  list: (page = 1, pageSize = 10, activityId) => 
    request.get('/admin/registrations', { page, pageSize, activityId }),
  review: (id, status) => 
    request.put(`/admin/registrations/${id}/review`, { status }),
  delete: (id) => 
    request.delete(`/admin/registrations/${id}`)
}

// 设置管理 API
export const settingsApi = {
  get: () => 
    request.get('/admin/settings'),
  save: (data) => 
    request.post('/admin/settings', data)
}

// 公司相关 API
export const companyApi = {
  list: (page = 1, pageSize = 100, keyword = '') =>
    request.get('/companies', { page, pageSize, keyword }),
  get: (id) =>
    request.get(`/companies/${id}`),
  create: (data) =>
    request.post('/companies', data),
  update: (id, data) =>
    request.put(`/companies/${id}`, data),
  delete: (id) =>
    request.delete(`/companies/${id}`)
}

// 分支机构相关 API
export const branchApi = {
  list: (page = 1, pageSize = 100, companyId = '', keyword = '') =>
    request.get('/branches', { page, pageSize, companyId, keyword }),
  get: (id) =>
    request.get(`/branches/${id}`),
  create: (data) =>
    request.post('/branches', data),
  update: (id, data) =>
    request.put(`/branches/${id}`, data),
  delete: (id) =>
    request.delete(`/branches/${id}`)
}

// 分享相关 API
export const shareApi = {
  generate: (data) =>
    request.post('/share/generate', data),
  records: (page = 1, pageSize = 20) =>
    request.get('/share/records', { page, pageSize }),
  visit: (shareId) =>
    request.get(`/visitor/activity/${shareId}`)
}

// 文件上传 API
export const uploadApi = {
  uploadImage: (filePath, filename) =>
    request.upload('/upload/image', filePath, filename),
  uploadFile: (filePath, filename) =>
    request.upload('/upload/file', filePath, filename)
}

// 扩展 request 对象添加 upload 方法
request.upload = function(url, filePath, filename) {
  return new Promise((resolve, reject) => {
    const app = getApp()
    const baseUrl = app.globalData.apiBaseUrl || 'http://192.168.8.230:8082/api/v1'
    const fullUrl = `${baseUrl}${url}`
    wx.uploadFile({
      url: fullUrl,
      filePath: filePath,
      name: 'file',
      header: {
        ...this.getAuthHeader()
      },
      success: (response) => {
        if (response.statusCode === 200) {
          try {
            const data = JSON.parse(response.data)
            if (data.success && data.data && data.data.url) {
              resolve(data.data.url)
            } else {
              reject(new Error(data.message || '上传失败'))
            }
          } catch (e) {
            reject(new Error('解析响应失败'))
          }
        } else {
          reject(new Error('上传失败'))
        }
      },
      fail: (error) => {
        console.error('Upload fail:', error)
        wx.showToast({ title: error.errMsg || '上传失败', icon: 'none' })
        reject(error)
      }
    })
  })
}
