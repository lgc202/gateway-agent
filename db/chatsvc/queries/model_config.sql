-- name: InsertModelConfig :execresult
INSERT INTO model_configs (name, provider, model, base_url, api_key, max_tokens)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetModelConfig :one
SELECT id,
       name,
       provider,
       model,
       base_url,
       max_tokens,
       api_key IS NOT NULL AS api_key_configured,
       created_at,
       updated_at
FROM model_configs
WHERE id = ?;

-- name: ListModelConfigs :many
SELECT id,
       name,
       provider,
       model,
       base_url,
       max_tokens,
       api_key IS NOT NULL AS api_key_configured,
       created_at,
       updated_at
FROM model_configs
ORDER BY id;

-- name: GetModelConfigAPIKey :one
SELECT api_key
FROM model_configs
WHERE id = ?;

-- name: UpdateModelConfig :execrows
UPDATE model_configs
SET name = ?,
    provider = ?,
    model = ?,
    base_url = ?,
    max_tokens = ?,
    updated_at = CURRENT_TIMESTAMP(6)
WHERE id = ?;

-- name: UpdateModelConfigAPIKey :execrows
UPDATE model_configs
SET api_key = ?,
    updated_at = CURRENT_TIMESTAMP(6)
WHERE id = ?;

-- name: DeleteModelConfig :execrows
DELETE
FROM model_configs
WHERE id = ?;

-- name: CountChatsByModelConfigID :one
SELECT COUNT(*)
FROM chats
WHERE model_config_id = ?;
