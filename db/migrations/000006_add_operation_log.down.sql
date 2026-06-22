DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/api/v1/logs' AND method = 'GET' AND name = '日志查询'
);
DELETE FROM `role_permission` WHERE `permission_id` IN (
    SELECT id FROM permission WHERE path = '/system/logs' AND method = 'MENU'
);
DELETE FROM `permission` WHERE path LIKE '/api/v1/logs%' OR path = '/system/logs';

DROP TABLE IF EXISTS `operation_log`;
