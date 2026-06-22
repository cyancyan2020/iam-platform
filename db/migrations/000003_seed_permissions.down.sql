DELETE FROM role_permission WHERE permission_id IN (
    SELECT id FROM permission WHERE path LIKE '/system/%' OR path LIKE '/api/v1/roles%' OR path LIKE '/api/v1/permissions%' OR path LIKE '/api/v1/users%' AND method NOT IN ('GET', 'POST') AND path = '/api/v1/profile'
);

DELETE FROM permission WHERE path LIKE '/system/%'
    OR (path LIKE '/api/v1/users%' AND method = 'GET' AND name = '用户列表')
    OR (path LIKE '/api/v1/users%' AND method = 'POST' AND name = '分配角色')
    OR (path LIKE '/api/v1/roles%' AND id NOT IN (SELECT id FROM (SELECT id FROM permission WHERE path = '/api/v1/profile') AS t))
    OR (path LIKE '/api/v1/permissions%');
