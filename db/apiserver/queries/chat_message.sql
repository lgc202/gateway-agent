-- name: InsertChatMessage :execresult
INSERT INTO chat_messages (chat_id, role, content)
VALUES (?, ?, ?);

-- name: GetChatMessage :one
SELECT id, chat_id, role, content, created_at
FROM chat_messages
WHERE id = ?;

-- name: ListChatMessagesAfter :many
SELECT id, chat_id, role, content, created_at
FROM chat_messages
WHERE chat_id = ?
  AND id > ?
ORDER BY id
LIMIT ?;
