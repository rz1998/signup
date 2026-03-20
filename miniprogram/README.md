# 报名项目微信小程序

微信小程序原生开发项目，使用 TypeScript。

## 项目结构

```
miniprogram/
├── app.js/app.ts           # 小程序入口
├── app.json                # 全局配置
├── app.wxss                # 全局样式
├── project.config.json     # 项目配置
├── sitemap.json            # 网站地图配置
├── tsconfig.json           # TypeScript 配置
├── utils/                  # 工具类
│   ├── typings.ts         # 类型定义
│   └── request.ts         # API 请求封装
├── assets/                 # 静态资源
│   └── icons/             # 图标
└── pages/                  # 页面
    ├── index/             # 首页（活动列表）
    ├── activity/          # 活动详情
    ├── my-registration/   # 我的报名
    ├── login/             # 登录页
    └── admin/             # 管理端（分包）
        ├── dashboard/     # 仪表盘
        ├── users/         # 用户管理
        ├── activities/    # 活动管理
        ├── registrations/ # 报名管理
        └── profile/       # 个人中心
```

## 页面说明

### 游客端（主包）
- **首页**：展示活动列表
- **活动详情**：查看活动并报名
- **我的报名**：查看已报名的活动

### 管理端（分包，需登录）
- **Dashboard**：统计面板
- **用户管理**：查看/删除用户
- **活动管理**：创建/编辑/删除活动
- **报名管理**：审核报名记录
- **个人中心**：修改个人信息

## 开发说明

1. 修改 `project.config.json` 中的 `appid` 为你的小程序 AppID
2. 在 `utils/request.ts` 中修改 `apiBaseUrl` 为后端 API 地址
3. 使用微信开发者工具打开项目目录

## API 基础地址

默认配置：`http://localhost:8080/api/v1`

可在 `app.ts` 的 `globalData.apiBaseUrl` 中修改。
