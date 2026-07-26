CREATE TABLE model_configs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模型配置唯一标识',
    name VARCHAR(64) NOT NULL COMMENT '模型配置显示名称',
    provider VARCHAR(32) NOT NULL COMMENT '模型服务提供方',
    model VARCHAR(255) NOT NULL COMMENT '模型服务中的模型名称',
    base_url VARCHAR(2048) NOT NULL DEFAULT '' COMMENT '模型服务地址，空字符串表示使用默认地址',
    api_key VARBINARY(2048) NULL COMMENT 'AES-256-GCM 加密后的模型 API Key',
    max_tokens INT UNSIGNED NOT NULL COMMENT '模型单次回复的最大 Token 数',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '模型配置创建时间',
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '模型配置更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_model_configs_name (name)
) ENGINE=InnoDB COMMENT='Gateway Agent 可用的模型配置';
