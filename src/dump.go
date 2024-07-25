package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/jackc/pgx/v5"
	"sasakiy84.net/relive-multi-aggregator/src/db"
)

func Dump(resultDir string) error {
	DATABASE_URL, ok := os.LookupEnv("POSTGRESQL_URL")
	if !ok {
		return fmt.Errorf("POSTGRESQL_URL is not set")
	}

	dbCtx := context.Background()
	conn, err := pgx.Connect(dbCtx, DATABASE_URL)
	if err != nil {
		return err
	}
	defer conn.Close(dbCtx)

	queries := db.New(conn)

	events, err := queries.ListEvents(dbCtx, nil)
	if err != nil {
		return err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	targetDir := path.Join(currentDir, resultDir)
	err = os.MkdirAll(targetDir, 0755)
	if err != nil {
		return err
	}

	// dump events as JSON "events.json"
	if eventJson, err := json.MarshalIndent(events, "", "  "); err != nil {
		return err
	} else {
		err := os.WriteFile(path.Join(targetDir, "events.json"), eventJson, 0644)
		if err != nil {
			return err
		}
	}

	// dump videos as JSON [event_name].json
	eventDir := path.Join(targetDir, "events")
	err = os.MkdirAll(eventDir, 0755)
	if err != nil {
		return err
	}

	for _, event := range events {

		videos, err := queries.GetVideosByEventId(dbCtx, db.GetVideosByEventIdParams{
			EventID: event.ID,
		})
		if err != nil {
			return err
		}

		for _, video := range videos {
			if jsonVideo, err := json.MarshalIndent(video, "", "  "); err != nil {
				return err
			} else {
				fmt.Printf("Video: %s", jsonVideo)
			}
		}

		videosJson, err := json.Marshal(videos)
		if err != nil {
			return err
		}

		eventId, err := event.ID.Value()
		if err != nil {
			return err
		}
		videosFilePath := path.Join(eventDir, eventId.(string)+".json")
		err = os.WriteFile(videosFilePath, videosJson, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
