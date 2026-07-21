CREATE TABLE chat_messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '消息唯一标识',
    chat_id BIGINT UNSIGNED NOT NULL COMMENT '消息所属对话',
    role VARCHAR(16) NOT NULL COMMENT '消息角色',
    content TEXT NOT NULL COMMENT '消息正文',
    created_at DATETIME(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) COMMENT '消息创建时间',
    PRIMARY KEY (id),
    INDEX idx_chat_messages_chat_id_id (chat_id, id)
) ENGINE=InnoDB COMMENT='对话中的消息';
