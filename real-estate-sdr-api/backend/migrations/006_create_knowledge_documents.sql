CREATE TABLE knowledge_documents (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    source VARCHAR(255),PS C:\Users\rodri\Documents\codes\fiap\8IADT-tech-challenge-fase-5\real-estate-sdr-api> Get-Content .\backend\queries\messages.sql
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
ORDER BY created_at DESC
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
PS C:\Users\rodri\Documents\codes\fiap\8IADT-tech-challenge-fase-5\real-estate-sdr-api> Get-Content .\backend\queries\messages.sql |
>>     Select-Object -Skip 35 -First 25
    content,
    created_at
FROM messages
ORDER BY created_at DESC
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


PS C:\Users\rodri\Documents\codes\fiap\8IADT-tech-challenge-fase-5\real-estate-sdr-api> 
    metadata JSONB,
    embedding VECTOR(1536),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_knowledge_documents_embedding
ON knowledge_documents
USING hnsw (embedding vector_cosine_ops);