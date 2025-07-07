package cronjob

import (
	"log"
	"time"

	"go-rfp/internal/csvsync"

	"github.com/robfig/cron/v3"
)

func StartScheduler() {
	c := cron.New()
	c.AddFunc("@every 24h", func() {
		log.Println("🔁 Starting daily CSV sync...")
		if err := csvsync.DownloadAndSync(); err != nil {
			log.Printf("❌ Sync failed: %v", err)
		} else {
			log.Println("✅ Sync completed successfully")
		}
	})
	c.Start()

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("📦 Initial CSV sync starting...")
		if err := csvsync.DownloadAndSync(); err != nil {
			log.Printf("❌ Initial sync failed: %v", err)
		}
	}()
}
