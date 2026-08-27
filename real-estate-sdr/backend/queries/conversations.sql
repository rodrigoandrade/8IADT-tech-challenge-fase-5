-- name: CreateConversation :one
INSERT INTO conversations (
    lead_id,
    status,
    started_at,
    ended_at
)
VALUES (
           $1,
           $2,
           CURRENT_TIMESTAMP,
           NULL
       )
    RETURNING
    id,
    lead_id,
    status,
    started_at,
    ended_at;


-- name: GetConversationByID :one
SELECT
    id,
    lead_id,
    status,
    started_at,
    ended_at
FROM conversations
WHERE id = $1;


-- name: ListConversationsByLead :many
SELECT
    id,
    lead_id,
    status,
    started_at,
    ended_at
FROM conversations
WHERE lead_id = $1
ORDER BY started_at DESC;


-- name: ListConversations :many
SELECT
    id,
    lead_id,
    status,
    started_at,
    ended_at
FROM conversations
ORDER BY started_at DESC
    LIMIT $1
OFFSET $2;


-- name: CloseConversation :one
UPDATE conversations
SET
    status = 'CLOSED',
    ended_at = CURRENT_TIMESTAMP
WHERE id = $1
  AND status = 'ACTIVE'
    RETURNING
    id,
    lead_id,
    status,
    started_at,
    ended_at;


-- name: CountConversations :one
SELECT COUNT(*)
FROM conversations;