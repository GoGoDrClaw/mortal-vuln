package tracker

import (
	"fmt"
	"log"
	"os"
	"time"
	"vulnnotes/middleware"
)

// ResetAllTeams resets progress for every team.
func ResetAllTeams() {
	for _, team := range middleware.Teams {
		ResetTeam(team)
	}
	log.Printf("🔄 All team progress reset (%d teams)", len(middleware.Teams))
}

// StartResetScheduler reads RESET_TIME (HH:MM) and RESET_TIMEZONE from env
// and runs a full reset (progress + DB) at that time every day.
// An optional dbReset callback lets handlers.SeedAllDBs be wired in from main.
func StartResetScheduler(dbReset func()) {
	resetTime := os.Getenv("RESET_TIME")
	if resetTime == "" {
		log.Println("⏰ RESET_TIME not set — scheduled reset disabled")
		return
	}

	tz := os.Getenv("RESET_TIMEZONE")
	if tz == "" {
		tz = "UTC"
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("⚠️  Invalid RESET_TIMEZONE=%q, falling back to UTC: %v", tz, err)
		loc = time.UTC
		tz = "UTC"
	}

	var hh, mm int
	if _, err := fmt.Sscanf(resetTime, "%d:%02d", &hh, &mm); err != nil || hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		log.Printf("⚠️  Invalid RESET_TIME=%q (want HH:MM) — scheduled reset disabled", resetTime)
		return
	}

	log.Printf("⏰ Scheduled reset at %02d:%02d %s every day", hh, mm, tz)

	go func() {
		for {
			now := time.Now().In(loc)
			next := time.Date(now.Year(), now.Month(), now.Day(), hh, mm, 0, 0, loc)
			if !next.After(now) {
				next = next.Add(24 * time.Hour)
			}

			log.Printf("⏰ Next auto-reset: %s", next.Format("2006-01-02 15:04:05 MST"))
			<-time.After(time.Until(next))

			log.Println("🔄 Scheduled reset triggered — resetting all teams...")
			ResetAllTeams()
			if dbReset != nil {
				dbReset()
			}
		}
	}()
}
