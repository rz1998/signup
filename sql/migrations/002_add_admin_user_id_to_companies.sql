-- 添加公司管理员字段到 companies 表
ALTER TABLE companies ADD COLUMN IF NOT EXISTS admin_user_id UUID REFERENCES users(id) ON DELETE SET NULL;

COMMENT ON COLUMN companies.admin_user_id IS '公司管理员用户ID';
