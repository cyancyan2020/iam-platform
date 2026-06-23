-- 创建默认管理员用户（admin / admin123）
INSERT INTO `user` (`username`, `password_hash`, `nickname`, `role_id`, `created_at`, `updated_at`)
VALUES ('admin', '$2a$10$Quyz73glj/0HBrIO53U1pe5tPO7w0nPTEL7SppoxmHAbEBRvCORNa', '管理员', 1, NOW(), NOW());
