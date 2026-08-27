-- name: CreateMessage :one
INSERT INTO messages (
    conversation_id,
    role,
    content
)
VALUES (
           $1,
           $2,
           $3
       )
    RETURNING
    id,
    conversation_id,
    role,
    content,
    created_at;


-- name: GetMessageByID :one
SELECT
    id,
    conversation_id,
    role,
    content,
    created_at
FROM messages
WHERE id = $1;


-- name: ListMessages :many
SELECT
    id,
    conversation_id,
    role,
    content,
    created_at
FROM messages
ORDER BY created_at ASC
    LIMIT $1
OFFSET $2;


-- name: ListMessagesByConversation :many
SELECT
    id,
    conversation_id,
    role,
    content,
    created_at
FROM messages
WHERE conversation_id = $1
ORDER BY created_at ASC;


-- name: CountMessages :one
SELECT COUNT(*)
FROM messages;


-- name: CountMessagesByConversation :one
SELECT COUNT(*)
FROM messages
WHERE conversation_id = $1;