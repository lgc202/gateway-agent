-- name: InsertChat :execresult
INSERT INTO chats (model_config_id)
VALUES (?);

-- name: GetChat :one
SELECT id, model_config_id, created_at, updated_at
FROM chats
WHERE id = ?;

-- name: GetChatForUpdate :one
SELECT id
FROM chats
WHERE id = ?
FOR UPDATE;

-- name: TouchChat :execrows
UPDATE chats
SET updated_at = GREATEST(CURRENT_TIMESTAMP(6), TIMESTAMPADD(MICROSECOND, 1, updated_at))
WHERE id = ?;
