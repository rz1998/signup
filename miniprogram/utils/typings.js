// 用户信息
export interface UserInfo {
  id: number
  username: string
  nickname: string
  avatar: string
  role: 'admin' | 'user'
  phone?: string
  email?: string
  createdAt: string
}

// 活动信息
export interface Activity {
  id: number
  title: string
  description: string
  coverImage: string
  startTime: string
  endTime: string
  location: string
  maxParticipants: number
  currentParticipants: number
  status: 'draft' | 'published' | 'ongoing' | 'ended' | 'cancelled'
  createdBy: number
  createdAt: string
  updatedAt: string
}

// 报名信息
export interface Registration {
  id: number
  activityId: number
  activityTitle: string
  userId: number
  username: string
  nickname: string
  phone: string
  status: 'pending' | 'confirmed' | 'cancelled' | 'completed'
  remark?: string
  createdAt: string
  updatedAt: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  token: string
  userInfo: UserInfo
}

// API 响应
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 分页响应
export interface PageResponse<T = any> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

// 报名请求
export interface RegistrationRequest {
  activityId: number
  phone: string
  remark?: string
}

// 小程序 App 选项
export interface IAppOption {
  globalData: {
    userInfo: UserInfo | null
    token: string
    apiBaseUrl: string
  }
}
