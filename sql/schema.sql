-- 报名系统数据库表结构 v2
-- 支持多租户架构：公司(companies) -> 分支机构(branches) -> 用户(users)
-- Database: signup_db

-- 公司表
CREATE TABLE IF NOT EXISTS companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active' NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_companies_name ON companies(name);
CREATE INDEX idx_companies_status ON companies(status);
CREATE INDEX idx_companies_deleted_at ON companies(deleted_at) WHERE deleted_at IS NULL;

-- 分支机构表
CREATE TABLE IF NOT EXISTS branches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    leader_name VARCHAR(50),
    leader_phone VARCHAR(20),
    leader_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_branches_company_id ON branches(company_id);
CREATE INDEX idx_branches_status ON branches(status);
CREATE INDEX idx_branches_deleted_at ON branches(deleted_at) WHERE deleted_at IS NULL;

-- 用户表（扩展支持租户）
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    email VARCHAR(100),
    phone VARCHAR(20),
    role VARCHAR(20) DEFAULT 'sales' NOT NULL CHECK (role IN ('admin', 'sys_mgr', 'company_mgr', 'branch_mgr', 'mgr', 'sales')),
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    manager_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'active' NOT NULL CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_branch_id ON users(branch_id);
CREATE INDEX idx_users_manager_id ON users(manager_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

-- 活动表（扩展支持公司层面和分支机构层面）
CREATE TABLE IF NOT EXISTS activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    cover_image VARCHAR(255),
    content JSONB,
    status VARCHAR(20) DEFAULT 'draft' NOT NULL CHECK (status IN ('draft', 'active', 'closed', 'cancelled', 'completed')),
    start_date DATE,
    end_date DATE,
    location VARCHAR(255),
    max_participants INTEGER DEFAULT 0,
    current_participants INTEGER DEFAULT 0,
    page_title VARCHAR(100),
    page_subtitle VARCHAR(200),
    page_description TEXT,
    registration_notice TEXT,
    contact_phone VARCHAR(20),
    contact_email VARCHAR(100),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_activities_company_id ON activities(company_id);
CREATE INDEX idx_activities_branch_id ON activities(branch_id);
CREATE INDEX idx_activities_name ON activities(name);
CREATE INDEX idx_activities_status ON activities(status);
CREATE INDEX idx_activities_created_by ON activities(created_by);
CREATE INDEX idx_activities_deleted_at ON activities(deleted_at) WHERE deleted_at IS NULL;

-- 表单字段表
CREATE TABLE IF NOT EXISTS form_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    field_name VARCHAR(50) NOT NULL,
    field_type VARCHAR(20) NOT NULL CHECK (field_type IN ('text', 'textarea', 'radio', 'checkbox', 'select', 'date', 'file')),
    is_required BOOLEAN DEFAULT FALSE,
    is_default BOOLEAN DEFAULT FALSE,
    options JSONB,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_form_fields_activity_id ON form_fields(activity_id);
CREATE INDEX idx_form_fields_deleted_at ON form_fields(deleted_at) WHERE deleted_at IS NULL;

-- 报名信息表（扩展支持分支机构）
CREATE TABLE IF NOT EXISTS registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    branch_id UUID REFERENCES branches(id) ON DELETE SET NULL,
    sales_id UUID REFERENCES users(id) ON DELETE SET NULL,
    share_id UUID,
    visitor_openid VARCHAR(100),
    form_data JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending' NOT NULL CHECK (status IN ('pending', 'confirmed', 'rejected', 'cancelled')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_registrations_activity_id ON registrations(activity_id);
CREATE INDEX idx_registrations_branch_id ON registrations(branch_id);
CREATE INDEX idx_registrations_sales_id ON registrations(sales_id);
CREATE INDEX idx_registrations_share_id ON registrations(share_id);
CREATE INDEX idx_registrations_visitor_openid ON registrations(visitor_openid);
CREATE INDEX idx_registrations_status ON registrations(status);
CREATE INDEX idx_registrations_deleted_at ON registrations(deleted_at) WHERE deleted_at IS NULL;

-- 分享记录表
CREATE TABLE IF NOT EXISTS shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_id UUID NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    sales_id UUID REFERENCES users(id) ON DELETE SET NULL,
    share_type VARCHAR(20) DEFAULT 'link' NOT NULL CHECK (share_type IN ('link', 'qrcode')),
    share_url TEXT,
    qr_code_data TEXT,
    visit_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_shares_activity_id ON shares(activity_id);
CREATE INDEX idx_shares_sales_id ON shares(sales_id);

-- 创建默认管理员用户 (密码: admin123)
INSERT INTO users (username, password, full_name, email, phone, role, status)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMye/p3F5GhR1L/1m5.r5HXrMkfqXZbneGi', '管理员', 'admin@example.com', '13800138000', 'admin', 'active')
ON CONFLICT (username) DO NOTHING;

-- 创建测试用户
INSERT INTO users (username, password, full_name, email, phone, role, status)
VALUES 
    ('manager1', '$2a$10$N9qo8uLOickgx2ZMRZoMye/p3F5GhR1L/1m5.r5HXrMkfqXZbneGi', '经理1', 'manager1@example.com', '13800138001', 'mgr', 'active'),
    ('sales1', '$2a$10$N9qo8uLOickgx2ZMRZoMye/p3F5GhR1L/1m5.r5HXrMkfqXZbneGi', '销售1', 'sales1@example.com', '13800138002', 'sales', 'active')
ON CONFLICT (username) DO NOTHING;
