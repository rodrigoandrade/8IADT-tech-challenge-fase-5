-- name: CreateProperty :one
INSERT INTO properties (
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available
)
VALUES (
           $1,
           $2,
           $3,
           $4,
           $5,
           $6,
           $7,
           $8,
           $9,
           $10,
           $11,
           $12,
           $13
       )
    RETURNING
    id,
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available,
    created_at,
    updated_at;


-- name: GetPropertyByID :one
SELECT
    id,
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available,
    created_at,
    updated_at
FROM properties
WHERE id = $1;


-- name: ListProperties :many
SELECT
    id,
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available,
    created_at,
    updated_at
FROM properties
ORDER BY created_at DESC
    LIMIT $1
OFFSET $2;


-- name: UpdateProperty :one
UPDATE properties
SET
    title = $2,
    description = $3,
    property_type = $4,
    transaction_type = $5,
    price = $6,
    bedrooms = $7,
    bathrooms = $8,
    parking_spaces = $9,
    area = $10,
    city = $11,
    neighborhood = $12,
    address = $13,
    available = $14,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    RETURNING
    id,
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available,
    created_at,
    updated_at;


-- name: UpdatePropertyAvailability :one
UPDATE properties
SET
    available = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
    RETURNING
    id,
    title,
    description,
    property_type,
    transaction_type,
    price,
    bedrooms,
    bathrooms,
    parking_spaces,
    area,
    city,
    neighborhood,
    address,
    available,
    created_at,
    updated_at;


-- name: CountProperties :one
SELECT COUNT(*)
FROM properties;