CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CONSTRAINT non_empty_name CHECK (name <> '') CONSTRAINT unique_name UNIQUE,
    description TEXT NOT NULL default '',
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE keywords (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    word VARCHAR(50) NOT NULL CONSTRAINT non_empty_word CHECK (word <> '') CONSTRAINT unique_word UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE event_keywords (
    event_id UUID,
    keyword_id UUID,
    PRIMARY KEY (event_id, keyword_id),
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE,
    FOREIGN KEY (keyword_id) REFERENCES keywords(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE channels (
    youtube_channel_id TEXT PRIMARY KEY CONSTRAINT non_empty_id CHECK (youtube_channel_id <> ''),
    name TEXT NOT NULL CONSTRAINT non_empty_title CHECK (name <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE videos (
    youtube_video_id TEXT PRIMARY KEY CONSTRAINT non_empty_id CHECK (youtube_video_id <> ''),
    title TEXT NOT NULL CONSTRAINT non_empty_title CHECK (title <> ''),
    actual_end_time TIMESTAMPTZ NOT NULL,
    actual_start_time TIMESTAMPTZ NOT NULL,
    view_count BIGINT NOT NULL CONSTRAINT non_negative_view_count CHECK (view_count >= 0),
    like_count BIGINT NOT NULL CONSTRAINT non_negative_like_count CHECK (like_count >= 0),
    duration_second INTEGER NOT NULL CONSTRAINT positive_duration_second CHECK (duration_second >= 0),
    youtube_channel_id TEXT NOT NULL,
    event_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT actual_start_time_before_actual_end_time CHECK (actual_start_time < actual_end_time),
    FOREIGN KEY (youtube_channel_id) REFERENCES channels(youtube_channel_id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);

CREATE TYPE thumbnail_type AS ENUM (
    'default',
    'medium',
    'high',
    'standard',
    'maxres'
);

CREATE TABLE thumbnails (
    url TEXT PRIMARY KEY,
    type thumbnail_type NOT NULL,
    width INTEGER NOT NULL CONSTRAINT positive_width CHECK (width > 0),
    height INTEGER NOT NULL CONSTRAINT positive_height CHECK (height > 0),
    youtube_video_id TEXT NOT NULL,
    FOREIGN KEY (youtube_video_id) REFERENCES videos(youtube_video_id) ON DELETE CASCADE,
    CONSTRAINT single_thumbnail_type_per_video UNIQUE (youtube_video_id, type)
);

CREATE INDEX idx_event_keywords_event_id ON event_keywords USING HASH (event_id);

CREATE INDEX idx_event_keywords_keyword_id ON event_keywords USING HASH (keyword_id);

CREATE INDEX idx_videos_event_id ON videos USING HASH (event_id);

CREATE INDEX idx_videos_youtube_channel_id ON videos USING HASH (youtube_channel_id);