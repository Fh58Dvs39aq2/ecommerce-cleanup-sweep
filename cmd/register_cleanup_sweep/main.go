package main

import (
	"fmt"
	"log"
	"os"

	infrai "ecommerce-cleanup-sweep"
)

func main() {
	client, err := infrai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	jobID, err := infrai.ScheduleCleanupSweep(client, os.Getenv("CLEANUP_WEBHOOK_URL"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("scheduled cleanup sweep:", jobID)
}
