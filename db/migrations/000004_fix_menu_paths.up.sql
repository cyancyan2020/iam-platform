-- 修正 Phase 8 中 menu 路径单复数不一致问题
UPDATE `permission` SET `path` = '/system/users'       WHERE `path` = '/system/user';
UPDATE `permission` SET `path` = '/system/roles'       WHERE `path` = '/system/role';
UPDATE `permission` SET `path` = '/system/permissions' WHERE `path` = '/system/permission';

-- 补全 Phase 8 新增接口的权限记录

-- 角色权限查询接口
SET @role_menu_id = (SELECT id FROM permission WHERE path = '/system/roles' AND method = 'MENU');
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/roles/:id/permissions', 'GET', '查看角色权限', @role_menu_id);

-- 用户管理 CRUD 接口（GET /users 已在 000003 中，其余需补全）
SET @user_menu_id = (SELECT id FROM permission WHERE path = '/system/users' AND method = 'MENU');
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/users',       'POST',   '创建用户', @user_menu_id),
('/api/v1/users/:id',   'PUT',    '编辑用户', @user_menu_id),
('/api/v1/users/:id',   'DELETE', '删除用户', @user_menu_id);

-- 为管理员角色分配所有新增权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 1, id FROM permission WHERE id NOT IN (
    SELECT permission_id FROM role_permission WHERE role_id = 1
);
