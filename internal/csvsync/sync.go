package csvsync

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go-rfp/internal/db"
)

const (
	domain = "Contract Opportunities/datagov"
	fname  = "ContractOpportunitiesFullCSV.csv"
	base   = "https://sam.gov/api/prod/fileextractservices/v1/api"
)

func DownloadAndSync() error {
	// Construct download URL with proper domain quoting
	parts := strings.Split(domain, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	encodedDomain := strings.Join(parts, "/")
	downloadURL := fmt.Sprintf("%s/download/%s/%s?privacy=Public", base, encodedDomain, fname)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}

	file, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read header row: %v", err)
	}

	idx := map[string]int{}
	for i, h := range headers {
		idx[h] = i
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("csv read error: %v", err)
		}

		statusRaw := rec[idx["Active"]]
		if strings.ToLower(strings.TrimSpace(statusRaw)) != "yes" {
			continue
		}

		c := struct {
			ID, Title, Desc, Status, NAICS, Type,
			Posted, Resp, Award, Office, Agency string
		}{
			ID:     rec[idx["NoticeId"]],
			Title:  rec[idx["Title"]],
			Desc:   rec[idx["Description"]],
			Status: "active",
			NAICS:  strings.TrimSpace(rec[idx["NaicsCode"]]),
			Type:   rec[idx["Type"]],
			Posted: rec[idx["PostedDate"]],
			Resp:   rec[idx["ResponseDeadLine"]],
			Award:  rec[idx["AwardDate"]],
			Office: rec[idx["Office"]],
			Agency: rec[idx["Department/Ind.Agency"]],
		}

		_, err = tx.Exec(`
			INSERT OR REPLACE INTO contracts (
				id, title, description, status, naics, type, posted_date,
				response_deadline, award_date, contracting_office, agency, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			c.ID, c.Title, c.Desc, c.Status, c.NAICS,
			c.Type, c.Posted, c.Resp, c.Award,
			c.Office, c.Agency, time.Now(),
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("db insert error: %v", err)
		}
	}

	if _, err := tx.Exec(`DELETE FROM contracts WHERE status != 'active'`); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
