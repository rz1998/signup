-- 迁移脚本: 添加分享链接字段
-- 为 shares 表添加 share_url 列

-- 检查列是否存在，不存在则添加
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shares' AND column_name = 'share_url'
    ) THEN
        ALTER TABLE shares ADD COLUMN share_url TEXT;
    END IF;
END $$;

-- 添加索引（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_shares_created_at'
    ) THEN
        CREATE INDEX idx_shares_created_at ON shares(created_at);
    END IF;
END $$;
