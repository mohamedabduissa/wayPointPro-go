package jobs

import (
	"WayPointPro/pkg/traffic"
	"fmt"
	"log"
	"time"
)

// ResetRequestLimit job
func ResetRequestLimit() {
	// Get the current date
	now := time.Now()
	var cache = traffic.NewCache()
	// Check if it's the first day of the month
	if now.Day() == 1 {
		query := `UPDATE access_tokens SET request_count = 0, reset_date = $1`
		_, err := cache.DB.Exec(cache.CTX, query, time.Now())
		if err != nil {
			log.Println("Failed to reset request counts:", err)
		}
		fmt.Println("Request counts have been reset successfully!")
	} else {
		fmt.Println("Not the first day of the month. No reset required.")
	}
}
