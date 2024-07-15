-- name: ListVideos :many
SELECT
    *
FROM
    videos
ORDER BY
    created_at DESC, youtube_video_id DESC
LIMIT
    coalesce(sqlc.narg('limit')::int, 10);

-- name: GetVideoById :one
SELECT
    videos.*,
    channels.name AS channel_name,
    thumbnails.url AS thumbnail_url,
    thumbnails.type AS thumbnail_type
FROM
    videos
JOIN
    channels
ON
    videos.youtube_channel_id = channels.youtube_channel_id
JOIN
    thumbnails
ON
    videos.youtube_video_id = thumbnails.youtube_video_id
WHERE
    videos.youtube_video_id = $1
ORDER BY
    CASE 
        WHEN thumbnails.type = 'default' THEN 1
        WHEN thumbnails.type = 'medium' THEN 2
        WHEN thumbnails.type = 'high' THEN 3
        WHEN thumbnails.type = 'standard' THEN 4
        WHEN thumbnails.type = 'maxres' THEN 5
        ELSE 99
    END
LIMIT
    1;

-- name: GetVideosByEventId :many
SELECT
    videos.*,
    channels.name AS channel_name,
    picked_thumbnails.url AS thumbnail_url,
    picked_thumbnails.type AS thumbnail_type,
    picked_thumbnails.width AS thumbnail_width,
    picked_thumbnails.height AS thumbnail_height
FROM
    videos
JOIN
    channels
ON
    videos.youtube_channel_id = channels.youtube_channel_id
LEFT JOIN LATERAL (
    SELECT
        *
    FROM
        thumbnails
    WHERE
        thumbnails.youtube_video_id = videos.youtube_video_id
    ORDER BY
        CASE 
            WHEN thumbnails.type = 'default' THEN 1
            WHEN thumbnails.type = 'medium' THEN 2
            WHEN thumbnails.type = 'high' THEN 3
            WHEN thumbnails.type = 'standard' THEN 4
            WHEN thumbnails.type = 'maxres' THEN 5
            ELSE 99
        END
    LIMIT
        1
) AS picked_thumbnails ON true
WHERE
    videos.event_id = $1
ORDER BY
    videos.actual_start_time DESC, videos.actual_end_time DESC,
    videos.youtube_video_id DESC
LIMIT
    coalesce(sqlc.narg('limit')::int, NULL);

-- name: GetOrCreateVideo :one
WITH inserted_video AS (
    INSERT INTO videos (
        youtube_video_id,
        title,
        actual_end_time,
        actual_start_time,
        view_count,
        like_count,
        duration_second,
        youtube_channel_id,
        event_id
    ) VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9
    )
    ON CONFLICT (youtube_video_id) DO NOTHING
    RETURNING
        *, true AS created
)
SELECT
    *
FROM
    inserted_video
UNION ALL
(SELECT
    *, false AS created
FROM
    videos
WHERE
    youtube_video_id = $1)
LIMIT
    1;

INSERT INTO videos (
    youtube_video_id,
    title,
    actual_end_time,
    actual_start_time,
    view_count,
    like_count,
    duration_second,
    youtube_channel_id,
    event_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
)
RETURNING
    *;

-- name: CreateThumbnail :one
INSERT INTO thumbnails (
    url,
    type,
    width,
    height,
    youtube_video_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING
    *;

-- name: CreateThumbnails :copyfrom
INSERT INTO thumbnails (
    url,
    type,
    width,
    height,
    youtube_video_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
);

-- name: UpsertThumbnail :one
INSERT INTO thumbnails (
    url,
    type,
    width,
    height,
    youtube_video_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (youtube_video_id, type) DO UPDATE
SET
    url = $1,
    width = $3,
    height = $4
RETURNING
    *;
