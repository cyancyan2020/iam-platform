-- 菜单权限
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/system/user',        'MENU', '用户管理', 0),
('/system/role',        'MENU', '角色管理', 0),
('/system/permission',  'MENU', '权限管理', 0);

-- 用户管理页面权限
SET @user_menu_id = (SELECT id FROM permission WHERE path = '/system/user' AND method = 'MENU');
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/users',           'GET',    '用户列表',   @user_menu_id),
('/api/v1/users/:id/role',  'POST',   '分配角色',   @user_menu_id);

-- 角色管理页面权限
SET @role_menu_id = (SELECT id FROM permission WHERE path = '/system/role' AND method = 'MENU');
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/roles',                         'GET',    '角色列表',       @role_menu_id),
('/api/v1/roles',                         'POST',   '创建角色',       @role_menu_id),
('/api/v1/roles/:id',                     'PUT',    '编辑角色',       @role_menu_id),
('/api/v1/roles/:id',                     'DELETE', '删除角色',       @role_menu_id),
('/api/v1/roles/:id/permissions',        'POST',   '分配权限',       @role_menu_id);

-- 权限管理页面权限
SET @perm_menu_id = (SELECT id FROM permission WHERE path = '/system/permission' AND method = 'MENU');
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/permissions',           'GET',    '权限列表',     @perm_menu_id),
('/api/v1/permissions',           'POST',   '创建权限',     @perm_menu_id),
('/api/v1/permissions/:id',       'PUT',    '编辑权限',     @perm_menu_id),
('/api/v1/permissions/:id',       'DELETE', '删除权限',     @perm_menu_id);

-- 为管理员角色分配所有新增权限
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 1, id FROM permission WHERE id NOT IN (
    SELECT permission_id FROM role_permission WHERE role_id = 1
);
