-- name: GetOrCreateChannelById :one
WITH inserted_channel AS (
    INSERT INTO channels (
        youtube_channel_id,
        name
    ) VALUES (
        $1,
        $2
    )
    ON CONFLICT (youtube_channel_id) DO NOTHING
    RETURNING
        *, true AS created
)
SELECT
    *
FROM
    inserted_channel
UNION ALL
(SELECT
    *, false AS created
FROM
    channels
WHERE
    youtube_channel_id = $1)
LIMIT
    1;