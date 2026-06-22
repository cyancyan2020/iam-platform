ALTER TABLE `role`
    ADD COLUMN `data_scope` VARCHAR(16) NOT NULL DEFAULT 'self' AFTER `name`;

UPDATE `role` SET `data_scope` = 'all' WHERE `code` = 'admin';
