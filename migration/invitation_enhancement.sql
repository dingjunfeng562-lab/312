-- 邀请码注册机制增强迁移脚本
-- 添加缺失字段：status, used_at, used_ip, expires_at, creator_id, creator_name, notes

-- MySQL 版本
ALTER TABLE invitation
    ADD COLUMN status VARCHAR(20) DEFAULT 'unused' COMMENT '状态: unused, used, disabled',
    ADD COLUMN used_at DATETIME NULL COMMENT '使用时间',
    ADD COLUMN used_ip VARCHAR(45) NULL COMMENT '使用者IP地址',
    ADD COLUMN expires_at DATETIME NULL COMMENT '过期时间',
    ADD COLUMN creator_id INT NULL COMMENT '创建者管理员ID',
    ADD COLUMN creator_name VARCHAR(255) DEFAULT 'system' COMMENT '创建者名称',
    ADD COLUMN notes TEXT NULL COMMENT '备注信息';

-- 为现有数据设置默认状态
UPDATE invitation SET status = CASE 
    WHEN used = TRUE THEN 'used'
    ELSE 'unused'
END WHERE status IS NULL OR status = '';

-- 添加索引以提高查询性能
CREATE INDEX idx_invitation_status ON invitation(status);
CREATE INDEX idx_invitation_expires_at ON invitation(expires_at);
CREATE INDEX idx_invitation_creator_id ON invitation(creator_id);

-- SQLite 版本（如果使用 SQLite，请使用以下语句）
-- 注意：SQLite 不支持在一条 ALTER TABLE 语句中添加多列
-- 需要分别执行以下语句：

-- ALTER TABLE invitation ADD COLUMN status VARCHAR(20) DEFAULT 'unused';
-- ALTER TABLE invitation ADD COLUMN used_at DATETIME NULL;
-- ALTER TABLE invitation ADD COLUMN used_ip VARCHAR(45) NULL;
-- ALTER TABLE invitation ADD COLUMN expires_at DATETIME NULL;
-- ALTER TABLE invitation ADD COLUMN creator_id INT NULL;
-- ALTER TABLE invitation ADD COLUMN creator_name VARCHAR(255) DEFAULT 'system';
-- ALTER TABLE invitation ADD COLUMN notes TEXT NULL;

-- UPDATE invitation SET status = CASE 
--     WHEN used = 1 THEN 'used'
--     ELSE 'unused'
-- END WHERE status IS NULL OR status = '';

-- CREATE INDEX idx_invitation_status ON invitation(status);
-- CREATE INDEX idx_invitation_expires_at ON invitation(expires_at);
-- CREATE INDEX idx_invitation_creator_id ON invitation(creator_id);
