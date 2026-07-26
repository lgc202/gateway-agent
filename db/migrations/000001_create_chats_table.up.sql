CREATE TABLE chats (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '对话唯一标识',
    model_config_id BIGINT UNSIGNED NULL COMMENT '当前对话选择的模型配置，NULL 表示使用系统默认配置',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '对话创建时间',
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '对话最后活跃时间',
    PRIMARY KEY (id),
    INDEX idx_chats_model_config_id (model_config_id)
) ENGINE=InnoDB COMMENT='用户与 Gateway Agent 的持续对话';
