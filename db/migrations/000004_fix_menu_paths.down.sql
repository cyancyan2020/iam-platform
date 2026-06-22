-- 回退路径修正
UPDATE `permission` SET `path` = '/system/user'       WHERE `path` = '/system/users';
UPDATE `permission` SET `path` = '/system/role'       WHERE `path` = '/system/roles';
UPDATE `permission` SET `path` = '/system/permission' WHERE `path` = '/system/permissions';

-- 删除本次新增的权限记录
DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/api/v1/users' AND method = 'POST' AND name = '创建用户'
);
DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/api/v1/users/:id' AND method = 'PUT' AND name = '编辑用户'
);
DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/api/v1/users/:id' AND method = 'DELETE' AND name = '删除用户'
);
DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/api/v1/roles/:id/permissions' AND method = 'GET' AND name = '查看角色权限'
);

DELETE FROM `permission` WHERE path = '/api/v1/users'   AND method = 'POST'   AND name = '创建用户';
DELETE FROM `permission` WHERE path = '/api/v1/users/:id' AND method = 'PUT'    AND name = '编辑用户';
DELETE FROM `permission` WHERE path = '/api/v1/users/:id' AND method = 'DELETE' AND name = '删除用户';
DELETE FROM `permission` WHERE path = '/api/v1/roles/:id/permissions' AND method = 'GET' AND name = '查看角色权限';
