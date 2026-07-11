-- ====================================
-- 邀请码功能扩展数据库迁移脚本
-- 日期：2026-07-08
-- 功能：使用记录详情 + 过期时间 + 超级管理员权限
-- ====================================

-- 1. 扩展 invitation 表
ALTER TABLE invitation 
ADD COLUMN expires_at DATETIME DEFAULT NULL COMMENT '过期时间（NULL=永久有效）',
ADD COLUMN creator_id BIGINT DEFAULT 0 COMMENT '创建者ID（0=系统）',
ADD COLUMN creator_name VARCHAR(100) DEFAULT 'system' COMMENT '创建者用户名',
ADD COLUMN notes VARCHAR(500) DEFAULT '' COMMENT '备注信息',
ADD COLUMN used_at DATETIME DEFAULT NULL COMMENT '实际使用时间',
ADD COLUMN used_ip VARCHAR(50) DEFAULT '' COMMENT '使用者IP地址',
ADD INDEX idx_expires_at (expires_at),
ADD INDEX idx_creator_id (creator_id);

-- 2. 扩展 auth 表（管理员权限等级）
ALTER TABLE auth 
ADD COLUMN admin_level INT DEFAULT 1 COMMENT '1=普通管理员, 2=超级管理员';

-- 3. 设置超级管理员
UPDATE auth SET admin_level = 2 WHERE username IN ('baishuwan', 'root');

-- 4. 为历史数据补充默认值
UPDATE invitation SET creator_name = 'system' WHERE creator_name = '' OR creator_name IS NULL;
