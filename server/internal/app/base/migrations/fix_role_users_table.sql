-- 修复 base_role_users 表结构
-- 用于解决 "column base_role_users.user_model_id does not exist" 错误

-- 1. 检查表是否存在
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'base_role_users') THEN
        RAISE NOTICE 'base_role_users 表已存在';
        
        -- 2. 检查是否存在 user_model_id 列
        IF EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_name = 'base_role_users' AND column_name = 'user_model_id'
        ) THEN
            RAISE NOTICE '发现 user_model_id 列，开始修复表结构...';
            
            -- 3. 删除旧的约束（如果存在）
            ALTER TABLE base_role_users DROP CONSTRAINT IF EXISTS base_role_users_pkey CASCADE;
            
            -- 4. 删除 user_model_id 列
            ALTER TABLE base_role_users DROP COLUMN IF EXISTS user_model_id;
            
            -- 5. 确保 role_id 和 user_id 列存在且不为空
            ALTER TABLE base_role_users ALTER COLUMN role_id SET NOT NULL;
            ALTER TABLE base_role_users ALTER COLUMN user_id SET NOT NULL;
            
            -- 6. 添加新的复合主键
            ALTER TABLE base_role_users ADD PRIMARY KEY (role_id, user_id);
            
            -- 7. 添加外键约束（如果不存在）
            ALTER TABLE base_role_users 
                ADD CONSTRAINT fk_base_role_users_role 
                FOREIGN KEY (role_id) REFERENCES base_roles(id) ON DELETE CASCADE;
            
            ALTER TABLE base_role_users 
                ADD CONSTRAINT fk_base_role_users_user 
                FOREIGN KEY (user_id) REFERENCES base_users(id) ON DELETE CASCADE;
            
            RAISE NOTICE 'base_role_users 表结构修复完成';
        ELSE
            -- 检查是否已经是正确的结构
            IF EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_name = 'base_role_users' AND column_name = 'role_id'
            ) AND EXISTS (
                SELECT 1 FROM information_schema.columns 
                WHERE table_name = 'base_role_users' AND column_name = 'user_id'
            ) THEN
                RAISE NOTICE 'base_role_users 表结构正确，无需修复';
            ELSE
                RAISE EXCEPTION 'base_role_users 表结构异常，缺少必要字段';
            END IF;
        END IF;
    ELSE
        RAISE NOTICE 'base_role_users 表不存在，将使用 migrations.go 创建';
    END IF;
END $$;

-- 验证表结构
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'base_role_users'
ORDER BY ordinal_position;

-- 查看约束
SELECT 
    tc.constraint_name, 
    tc.constraint_type,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints tc
LEFT JOIN information_schema.key_column_usage kcu 
    ON tc.constraint_name = kcu.constraint_name
LEFT JOIN information_schema.constraint_column_usage ccu 
    ON tc.constraint_name = ccu.constraint_name
WHERE tc.table_name = 'base_role_users';
