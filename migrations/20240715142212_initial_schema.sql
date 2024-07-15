-- Create enum type "thumbnail_type"
CREATE TYPE "thumbnail_type" AS ENUM ('default', 'medium', 'high', 'standard', 'maxres');
-- Create "events" table
CREATE TABLE "events" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "name" text NOT NULL, "description" text NOT NULL DEFAULT '', "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("id"), CONSTRAINT "unique_name" UNIQUE ("name"), CONSTRAINT "non_empty_name" CHECK (name <> ''::text));
-- Create "keywords" table
CREATE TABLE "keywords" ("id" uuid NOT NULL DEFAULT gen_random_uuid(), "word" character varying(50) NOT NULL, "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("id"), CONSTRAINT "unique_word" UNIQUE ("word"), CONSTRAINT "non_empty_word" CHECK ((word)::text <> ''::text));
-- Create "event_keywords" table
CREATE TABLE "event_keywords" ("event_id" uuid NOT NULL, "keyword_id" uuid NOT NULL, "created_at" timestamptz NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("event_id", "keyword_id"), CONSTRAINT "event_keywords_event_id_fkey" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "event_keywords_keyword_id_fkey" FOREIGN KEY ("keyword_id") REFERENCES "keywords" ("id") ON UPDATE NO ACTION ON DELETE CASCADE);
-- Create index "idx_event_keywords_event_id" to table: "event_keywords"
CREATE INDEX "idx_event_keywords_event_id" ON "event_keywords" USING hash ("event_id");
-- Create index "idx_event_keywords_keyword_id" to table: "event_keywords"
CREATE INDEX "idx_event_keywords_keyword_id" ON "event_keywords" USING hash ("keyword_id");
-- Create "channels" table
CREATE TABLE "channels" ("youtube_channel_id" text NOT NULL, "name" text NOT NULL, "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("youtube_channel_id"), CONSTRAINT "non_empty_id" CHECK (youtube_channel_id <> ''::text), CONSTRAINT "non_empty_title" CHECK (name <> ''::text));
-- Create "videos" table
CREATE TABLE "videos" ("youtube_video_id" text NOT NULL, "title" text NOT NULL, "actual_end_time" timestamptz NOT NULL, "actual_start_time" timestamptz NOT NULL, "view_count" bigint NOT NULL, "like_count" bigint NOT NULL, "duration_second" integer NOT NULL, "youtube_channel_id" text NOT NULL, "event_id" uuid NULL, "created_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, "updated_at" timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY ("youtube_video_id"), CONSTRAINT "videos_event_id_fkey" FOREIGN KEY ("event_id") REFERENCES "events" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "videos_youtube_channel_id_fkey" FOREIGN KEY ("youtube_channel_id") REFERENCES "channels" ("youtube_channel_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "actual_start_time_before_actual_end_time" CHECK (actual_start_time < actual_end_time), CONSTRAINT "non_empty_id" CHECK (youtube_video_id <> ''::text), CONSTRAINT "non_empty_title" CHECK (title <> ''::text), CONSTRAINT "non_negative_like_count" CHECK (like_count >= 0), CONSTRAINT "non_negative_view_count" CHECK (view_count >= 0), CONSTRAINT "positive_duration_second" CHECK (duration_second >= 0));
-- Create index "idx_videos_event_id" to table: "videos"
CREATE INDEX "idx_videos_event_id" ON "videos" USING hash ("event_id");
-- Create index "idx_videos_youtube_channel_id" to table: "videos"
CREATE INDEX "idx_videos_youtube_channel_id" ON "videos" USING hash ("youtube_channel_id");
-- Create "thumbnails" table
CREATE TABLE "thumbnails" ("url" text NOT NULL, "type" "thumbnail_type" NOT NULL, "width" integer NOT NULL, "height" integer NOT NULL, "youtube_video_id" text NOT NULL, PRIMARY KEY ("url"), CONSTRAINT "single_thumbnail_type_per_video" UNIQUE ("youtube_video_id", "type"), CONSTRAINT "thumbnails_youtube_video_id_fkey" FOREIGN KEY ("youtube_video_id") REFERENCES "videos" ("youtube_video_id") ON UPDATE NO ACTION ON DELETE CASCADE, CONSTRAINT "positive_height" CHECK (height > 0), CONSTRAINT "positive_width" CHECK (width > 0));
