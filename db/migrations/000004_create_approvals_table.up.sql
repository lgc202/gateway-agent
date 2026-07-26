CREATE TABLE approvals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '审批唯一标识',
    chat_id BIGINT UNSIGNED NOT NULL COMMENT '发起本次审批的对话',
    status VARCHAR(16) NOT NULL COMMENT '审批状态：pending、approved 或 rejected',
    operation VARCHAR(64) NOT NULL COMMENT '等待执行的 Tool 操作名称，例如 create_route',
    arguments JSON NOT NULL COMMENT '模型提交并展示给用户的完整 Tool 参数',
    resume_target VARCHAR(512) NOT NULL COMMENT '需要恢复的写操作调用标识',
    runtime_state MEDIUMBLOB NOT NULL COMMENT '恢复被暂停的 Agent 执行所需的运行状态',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '审批创建时间',
    updated_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '审批决定时间，未决定时等于创建时间',
    PRIMARY KEY (id),
    KEY idx_approvals_chat_status (chat_id, status, id)
) ENGINE=InnoDB COMMENT='等待用户决定的 Agent 写操作';
