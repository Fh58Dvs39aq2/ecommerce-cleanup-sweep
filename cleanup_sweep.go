package infrai

import "fmt"

// ScheduleCleanupSweep registers the daily callback that removes stale commerce records.
func ScheduleCleanupSweep(client *Client, taskURL string) (string, error) {
	if taskURL == "" {
		return "", fmt.Errorf("CLEANUP_WEBHOOK_URL is required")
	}
	job, err := client.CreateCron("15 2 * * *", taskURL)
	if err != nil {
		return "", fmt.Errorf("schedule cleanup sweep: %w", err)
	}
	return job.JobID, nil
}
