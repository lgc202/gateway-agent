-- name: InsertApproval :execresult
INSERT INTO approvals (chat_id, status, operation, arguments, resume_target, runtime_state)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetApproval :one
SELECT id,
       chat_id,
       status,
       operation,
       arguments,
       resume_target,
       runtime_state,
       created_at,
       updated_at
FROM approvals
WHERE id = ?
  AND chat_id = ?;

-- name: ListPendingApprovals :many
SELECT id,
       chat_id,
       status,
       operation,
       arguments,
       created_at,
       updated_at
FROM approvals
WHERE chat_id = ?
  AND status = 'pending'
ORDER BY id;

-- name: DecideApproval :execrows
UPDATE approvals
SET status = ?,
    updated_at = CURRENT_TIMESTAMP(6)
WHERE id = ?
  AND chat_id = ?
  AND status = 'pending';
