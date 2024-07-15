-- name: GetEventById :one
SELECT
    *
FROM
    events
WHERE
    id = $1
LIMIT
    1;

-- name: GetEventByName :one
SELECT
    *
FROM
    events
WHERE
    name = $1
LIMIT
    1;

-- name: GetEventsByIds :many
SELECT
    *
FROM
    events
WHERE
    id = ANY(@ids::text[])
ORDER BY
    created_at DESC, id DESC
LIMIT
    coalesce(sqlc.narg('limit')::int, 10);

-- name: ListEvents :many
SELECT
    *
FROM
    events
ORDER BY
    created_at DESC, id DESC
LIMIT
    coalesce(sqlc.narg('limit')::int, 10);

-- name: CreateEvent :one
INSERT INTO events (
    name,
    description
) VALUES (
    $1,
    coalesce(sqlc.narg(description)::text, '')
)
RETURNING
    *;

-- name: GetOrCreateEventById :one
WITH inserted_event AS (
    INSERT INTO events (
        name
    ) VALUES (
        $1
    )
    ON CONFLICT (name) DO NOTHING
    RETURNING
        *, true AS created
)
(SELECT * FROM inserted_event)
UNION ALL
(SELECT *, false AS created FROM events WHERE name = $1)
LIMIT 1;