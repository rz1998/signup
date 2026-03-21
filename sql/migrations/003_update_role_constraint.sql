-- 更新用户角色检查约束，添加 company_admin 和 branch_admin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role::text = ANY (ARRAY['admin'::character varying, 'company_admin'::character varying, 'branch_admin'::character varying, 'mgr'::character varying, 'sales'::character varying]::text[]));
