package main

import (
	"flag"
	"log"
	"os"
	"time"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <subcommand> [args]", os.Args[0])
	}
	subcommand := os.Args[1]

	retrieveCmd := flag.NewFlagSet("retrieve", flag.ExitOnError)
	retrieveActulaStartTimeBefore := retrieveCmd.String("start-before", "", "include video only if actual start time is before specified time")
	retrieveActulaStartTimeAfter := retrieveCmd.String("start-after", "", "include video only if actual start time is after specified time")

	switch subcommand {
	case "retrieve":
		if len(os.Args) < 4 {
			log.Fatalf("Usage: %s retrieve <event_name> <playlist_id>", os.Args[0])
		}
		eventName := os.Args[2]
		playlistId := os.Args[3]

		retrieveCmd.Parse(os.Args[4:])
		opt := &RetrieveOptions{}
		if startBefore, err := time.Parse(time.RFC3339, *retrieveActulaStartTimeBefore); err != nil {
			if *retrieveActulaStartTimeBefore != "" {
				log.Fatalf("Error parsing time: %v", err)
			}
		} else {
			opt.startBefore = &startBefore
		}

		if startAfter, err := time.Parse(time.RFC3339, *retrieveActulaStartTimeAfter); err != nil {
			if *retrieveActulaStartTimeAfter != "" {
				log.Fatalf("Error parsing time: %v", err)
			}
		} else {
			opt.startAfter = &startAfter
		}

		println("playlistId", playlistId)

		Retreive(eventName, playlistId, *opt)
	case "dump":
		if len(os.Args) < 3 {
			log.Fatalf("Usage: %s dump <result_dir>", os.Args[0])
		}
		resultDir := os.Args[2]
		err := Dump(resultDir)
		if err != nil {
			log.Fatalf("Error dumping data: %v", err)
		}
	default:
		log.Fatalf("Unknown subcommand: %s", subcommand)
	}

}
