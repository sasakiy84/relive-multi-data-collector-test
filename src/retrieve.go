package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"sasakiy84.net/relive-multi-aggregator/src/db"
)

const YOUTUBE_MAX_PAGE_SIZE = 50

type RetrieveOptions struct {
	startBefore *time.Time
	startAfter  *time.Time
}

func handleFatalPgError(err error, ignoredPgErrorCodes []string) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		for _, code := range ignoredPgErrorCodes {
			if pgErr.Code == code {
				log.Printf("Error pg context: %v (code = %s) %s", pgErr.Message, pgErr.Code, pgErr.Detail)
				return pgErr.Code
			}
		}
		log.Fatalf("Error pg context: %v (code = %s) %s", pgErr.Message, pgErr.Code, pgErr.Detail)
	}
	log.Fatalf("Error pg context: %v", err)
	return "" // unreachable
}

// becasese time.ParseDuration does not support ISO8601 duration
func parseISO8601Duration(duration string) (time.Duration, error) {
	result := time.Duration(0)

	for len(duration) > 0 {
		// 先頭が数字ではないときは、duration から取り除く
		if duration[0] < '0' || duration[0] > '9' {
			duration = duration[1:]
			continue
		}

		valueStr := ""
		for len(duration) > 0 && duration[0] >= '0' && duration[0] <= '9' {
			valueStr += string(duration[0])
			duration = duration[1:]
		}
		value, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			return result, fmt.Errorf("error parsing value: %v", err)
		}
		unit := string([]rune(duration)[0:1])

		switch unit {
		case "D":
			result += time.Duration(value) * 24 * time.Hour
		case "H":
			result += time.Duration(value) * time.Hour
		case "M":
			result += time.Duration(value) * time.Minute
		case "S":
			result += time.Duration(value) * time.Second
		default:
			return result, fmt.Errorf("unknown unit: %s", unit)
		}

		duration = duration[1:]
	}

	return result, nil
}

func createOrGetEventFromYoutubeVideo(queries *db.Queries, dbCtx context.Context, youtubeVideo *youtube.Video, event db.Event) {
	channel, err := queries.GetOrCreateChannelById(dbCtx, db.GetOrCreateChannelByIdParams{
		YoutubeChannelID: youtubeVideo.Snippet.ChannelId,
		Name:             youtubeVideo.Snippet.ChannelTitle,
	})
	if err != nil {
		handleFatalPgError(err, []string{})
	}
	if jsonChannel, err := json.MarshalIndent(channel, "", "  "); err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	} else {
		log.Printf("Channel: %s", jsonChannel)
	}

	actualStartTime, err := time.Parse(time.RFC3339, youtubeVideo.LiveStreamingDetails.ActualStartTime)
	if err != nil {
		log.Fatalf("Error parsing time: %v", err)
	}
	actualEndTime, err := time.Parse(time.RFC3339, youtubeVideo.LiveStreamingDetails.ActualEndTime)
	if err != nil {
		log.Fatalf("Error parsing time: %v", err)
	}
	duration, err := parseISO8601Duration(youtubeVideo.ContentDetails.Duration)
	if err != nil {
		log.Fatalf("Error parsing duration: %v", err)
	}
	durationSeconds := int32(duration.Seconds())

	video, err := queries.GetOrCreateVideo(dbCtx, db.GetOrCreateVideoParams{
		YoutubeVideoID:   youtubeVideo.Id,
		YoutubeChannelID: channel.YoutubeChannelID,
		Title:            youtubeVideo.Snippet.Title,
		ActualEndTime: pgtype.Timestamptz{
			Time:  actualEndTime,
			Valid: true,
		},
		ActualStartTime: pgtype.Timestamptz{
			Time:  actualStartTime,
			Valid: true,
		},
		ViewCount:      int64(youtubeVideo.Statistics.ViewCount),
		LikeCount:      int64(youtubeVideo.Statistics.LikeCount),
		DurationSecond: durationSeconds,
		EventID:        event.ID,
	})

	if err != nil {
		handleFatalPgError(err, []string{pgerrcode.UniqueViolation})
	} else {
		if jsonVideo, err := json.MarshalIndent(video, "", "  "); err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		} else {
			log.Printf("Video: %s", jsonVideo)
		}
	}

	// thumbnail の保存
	if youtubeVideo.Snippet.Thumbnails != nil {
		_, err = queries.UpsertThumbnail(dbCtx, db.UpsertThumbnailParams{
			YoutubeVideoID: video.YoutubeVideoID,
			Url:            youtubeVideo.Snippet.Thumbnails.Default.Url,
			Type:           "default",
			Width:          int32(youtubeVideo.Snippet.Thumbnails.Default.Width),
			Height:         int32(youtubeVideo.Snippet.Thumbnails.Default.Height),
		})
		if err != nil {
			handleFatalPgError(err, []string{})
		}
	}

	if youtubeVideo.Snippet.Thumbnails.Medium != nil {
		_, err = queries.UpsertThumbnail(dbCtx, db.UpsertThumbnailParams{
			YoutubeVideoID: video.YoutubeVideoID,
			Url:            youtubeVideo.Snippet.Thumbnails.Medium.Url,
			Type:           "medium",
			Width:          int32(youtubeVideo.Snippet.Thumbnails.Medium.Width),
			Height:         int32(youtubeVideo.Snippet.Thumbnails.Medium.Height),
		})

		if err != nil {
			handleFatalPgError(err, []string{})
		}
	}

	if youtubeVideo.Snippet.Thumbnails.High != nil {
		_, err = queries.UpsertThumbnail(dbCtx, db.UpsertThumbnailParams{
			YoutubeVideoID: video.YoutubeVideoID,
			Url:            youtubeVideo.Snippet.Thumbnails.High.Url,
			Type:           "high",
			Width:          int32(youtubeVideo.Snippet.Thumbnails.High.Width),
			Height:         int32(youtubeVideo.Snippet.Thumbnails.High.Height),
		})
		if err != nil {
			handleFatalPgError(err, []string{})
		}
	}

	if youtubeVideo.Snippet.Thumbnails.Standard != nil {
		_, err = queries.UpsertThumbnail(dbCtx, db.UpsertThumbnailParams{
			YoutubeVideoID: video.YoutubeVideoID,
			Url:            youtubeVideo.Snippet.Thumbnails.Standard.Url,
			Type:           "standard",
			Width:          int32(youtubeVideo.Snippet.Thumbnails.Standard.Width),
			Height:         int32(youtubeVideo.Snippet.Thumbnails.Standard.Height),
		})
		if err != nil {
			handleFatalPgError(err, []string{})
		}
	}

	if youtubeVideo.Snippet.Thumbnails.Maxres != nil {
		_, err = queries.UpsertThumbnail(dbCtx, db.UpsertThumbnailParams{
			YoutubeVideoID: video.YoutubeVideoID,
			Url:            youtubeVideo.Snippet.Thumbnails.Maxres.Url,
			Type:           "maxres",
			Width:          int32(youtubeVideo.Snippet.Thumbnails.Maxres.Width),
			Height:         int32(youtubeVideo.Snippet.Thumbnails.Maxres.Height),
		})
		if err != nil {
			handleFatalPgError(err, []string{})
		}
	}

}

func Retreive(eventName string, playlistId string, options RetrieveOptions) {
	API_KEY, ok := os.LookupEnv("YOUTUBE_API_KEY")
	if !ok {
		log.Fatalf("YOUTUBE_API is not set")
	}

	DATABASE_URL, ok := os.LookupEnv("POSTGRES_URL")
	if !ok {
		log.Fatalf("POSTGRES_URL is not set")
	}

	// databese の準備
	dbCtx := context.Background()
	conn, err := pgx.Connect(dbCtx, DATABASE_URL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
		os.Exit(1)
	}
	defer conn.Close(dbCtx)
	queries := db.New(conn)

	// event の設定
	events, err := queries.ListEvents(dbCtx, nil)
	if err != nil {
		log.Fatalf("Error getting event: %v", err)
	}
	for _, event := range events {
		if jsonEvent, err := json.MarshalIndent(event, "", "  "); err != nil {
			log.Fatalf("Error marshalling JSON: %v", err)
		} else {
			log.Printf("Event: %s", jsonEvent)
		}
	}
	targetEvent, err := queries.GetOrCreateEventById(dbCtx, eventName)
	if err != nil {
		handleFatalPgError(err, []string{})
	}

	// YouTube API の設定
	ctx := context.Background()
	youtubeService, err := youtube.NewService(ctx, option.WithAPIKey(API_KEY))
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// playlist から video を取得
	pagerToken := ""
	playlistsCaller := youtubeService.PlaylistItems.List([]string{
		"id",
		"snippet",
		"contentDetails",
	}).PlaylistId(playlistId).MaxResults(YOUTUBE_MAX_PAGE_SIZE)

	videoIds := []string{}
	for {
		response, err := playlistsCaller.Do()
		if err != nil {
			log.Fatalf("Error making search API call: %v", err)
		}

		videoIdsInPage := make([]string, len(response.Items))
		for i, item := range response.Items {
			fmt.Printf("Item: %s ; %s\n", item.Snippet.ResourceId.VideoId, item.Snippet.Title)
			videoIdsInPage[i] = item.Snippet.ResourceId.VideoId
		}
		videoIds = append(videoIds, videoIdsInPage...)

		fmt.Printf("Retrieved %d video IDs\n", len(videoIds))

		pagerToken = response.NextPageToken
		if int64(len(videoIds)) >= response.PageInfo.TotalResults {
			break
		}
		playlistsCaller.PageToken(pagerToken)
	}

	fmt.Printf("Total %d video IDs\n", len(videoIds))

	// playlist から取得した video から event を作成
	currentPage := 0
	for currentPage*YOUTUBE_MAX_PAGE_SIZE < len(videoIds) {
		start := currentPage * YOUTUBE_MAX_PAGE_SIZE
		end := (currentPage + 1) * YOUTUBE_MAX_PAGE_SIZE

		videoCaller := youtubeService.Videos.List([]string{
			"id",
			"snippet",
			"contentDetails",
			"liveStreamingDetails",
			"statistics",
			"recordingDetails",
			"liveStreamingDetails",
		}).Id(videoIds[start:end]...)

		response, err := videoCaller.Do()
		if err != nil {
			log.Fatalf("Error making search API call: %v", err)
		}

		for _, item := range response.Items {

			if options.startBefore != nil {
				actualStartTime, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
				if err != nil {
					log.Fatalf("Error parsing time: %v", err)
				}
				if actualStartTime.After(*options.startBefore) {
					continue
				}
			}

			if options.startAfter != nil {
				actualStartTime, err := time.Parse(time.RFC3339, item.LiveStreamingDetails.ActualStartTime)
				if err != nil {
					log.Fatalf("Error parsing time: %v", err)
				}
				if actualStartTime.Before(*options.startAfter) {
					continue
				}
			}

			log.Printf("ID: %v", item.Id)
			log.Printf("Title: %v", item.Snippet.Title)

			createOrGetEventFromYoutubeVideo(queries, dbCtx, item, db.Event{
				Name: eventName,
				ID:   targetEvent.ID,
			})

		}

		currentPage++
	}
}
