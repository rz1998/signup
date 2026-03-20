# 报名系统后端 API

基于 Go + go-zero + PostgreSQL 的报名系统后端服务。

## 功能特性

- 用户管理（管理员、营销人员、销售）
- 活动管理（创建、编辑、删除、状态管理）
- 表单字段配置（自定义报名字段）
- 报名信息管理（创建、编辑、审核）
- 分享功能（生成链接/二维码、来源追踪）

## 技术栈

- Go 1.22
- go-zero (REST 框架)
- PostgreSQL 15
- JWT 认证

## API 文档

启动服务后访问: `http://localhost:8082/api/docs`

### 认证 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/auth/login | 用户登录 |
| GET | /api/v1/auth/info | 获取当前用户 |
| POST | /api/v1/auth/logout | 登出 |

### 用户管理 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/users | 用户列表 |
| POST | /api/v1/users | 创建用户 |
| GET | /api/v1/users/:id | 获取用户 |
| PUT | /api/v1/users/:id | 更新用户 |
| DELETE | /api/v1/users/:id | 删除用户 |

### 活动管理 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/activities | 活动列表 |
| POST | /api/v1/activities | 创建活动 |
| GET | /api/v1/activities/:id | 获取活动 |
| PUT | /api/v1/activities/:id | 更新活动 |
| DELETE | /api/v1/activities/:id | 删除活动 |
| PUT | /api/v1/activities/:id/status | 更新活动状态 |

### 表单字段 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/activities/:id/fields | 获取表单字段 |
| POST | /api/v1/activities/:id/fields | 创建字段 |
| PUT | /api/v1/fields/:id | 更新字段 |
| DELETE | /api/v1/fields/:id | 删除字段 |

### 报名 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/activities/:id/registrations | 报名列表 |
| GET | /api/v1/registrations | 报名列表（全局） |
| POST | /api/v1/registrations | 创建报名 |
| GET | /api/v1/registrations/:id | 获取报名 |
| PUT | /api/v1/registrations/:id | 更新报名 |

### 分享 API

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/share/generate | 生成分享链接 |
| GET | /api/v1/visitor/activity/:shareId | 游客访问活动 |

## 快速开始

### 使用 Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f signup-api

# 停止服务
docker-compose down
```

### 手动启动

```bash
# 安装依赖
go mod download

# 运行
go run main.go -f etc/api.yaml

# 或使用配置文件
./signup-api -f etc/api.yaml
```

## 配置

配置文件: `etc/api.yaml`

```yaml
Name: signup-api
Host: 0.0.0.0
Port: 8082

Database:
  Type: postgres
  Host: fund-postgres
  Port: 5432
  DBName: signup_db
  User: signup_user
  Password: signup_password

JWT:
  Secret: your-secret-key
  Expire: 3600
```

## 默认账号

- 用户名: `admin`
- 密码: `admin123`
- 角色: `admin`

## 数据库

数据库初始化脚本: `sql/schema.sql`

表结构:
- `users` - 用户表
- `activities` - 活动表
- `form_fields` - 表单字段表
- `registrations` - 报名信息表
- `shares` - 分享记录表
