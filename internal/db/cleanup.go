package db

import (
	"log"
	"time"
)

// RemoveExpiredContracts deletes contracts where response_deadline < today
func RemoveExpiredContracts() {
	today := time.Now().Format("2006-01-02")

	query := `
		DELETE FROM contracts
		WHERE response_deadline < ?
	`

	res, err := DB.Exec(query, today)
	if err != nil {
		log.Printf("❌ Failed to remove expired contracts: %v", err)
		return
	}

	rowsDeleted, _ := res.RowsAffected()
	log.Printf("🧹 Cleaned up %d expired contracts", rowsDeleted)
}
