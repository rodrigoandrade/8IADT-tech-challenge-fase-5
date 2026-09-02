-- name: CreateLead :one
INSERT INTO leads (
    name,
    email,
    phone
)
VALUES (
           $1,
           $2,
           $3
       )
    RETURNING
    id,
    name,
    email,
    phone,
    status,
    created_at,
    updated_at;


-- name: GetLeadByID :one
SELECT
    id,
    name,
    email,
    phone,
    status,
    created_at,
    updated_at
FROM leads
WHERE id = $1;


-- name: ListLeads :many
SELECT
    id,
    name,
    email,
    phone,
    status,
    created_at,
    updated_at
FROM leads
ORDER BY created_at DESC;


-- name: UpdateLead :one
UPDATE leads
SET
    name = $2,
    email = $3,
    phone = $4,
    status = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    RETURNING
    id,
    name,
    email,
    phone,
    status,
    created_at,
    updated_at;


-- name: UpdateLeadStatus :one
UPDATE leads
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    RETURNING
    id,
    name,
    email,
    phone,
    status,
    created_at,
    updated_at;


-- name: CountLeads :one
SELECT COUNT(*)
FROM leads;