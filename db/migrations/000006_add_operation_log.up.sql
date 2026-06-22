CREATE TABLE IF NOT EXISTS `operation_log` (
    `id` BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `username` VARCHAR(64) NOT NULL,
    `method` VARCHAR(10) NOT NULL,
    `path` VARCHAR(256) NOT NULL,
    `ip` VARCHAR(45) NOT NULL,
    `user_agent` VARCHAR(512) DEFAULT '',
    `status_code` INT NOT NULL,
    `duration_ms` INT NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 操作日志菜单 + API 权限
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/system/logs', 'MENU', '操作日志', 0);

SET @log_menu_id = LAST_INSERT_ID();
INSERT INTO `permission` (`path`, `method`, `name`, `parent_id`) VALUES
('/api/v1/logs', 'GET', '日志查询', @log_menu_id);

INSERT INTO `role_permission` (`role_id`, `permission_id`) VALUES (1, @log_menu_id);
INSERT INTO `role_permission` (`role_id`, `permission_id`)
SELECT 1, id FROM permission WHERE id = LAST_INSERT_ID();
