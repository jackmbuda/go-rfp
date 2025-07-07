package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"go-rfp/internal/cronjob"
	"go-rfp/internal/db"
	"go-rfp/internal/openai"
)

type Contract struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Status            string `json:"status"`
	NAICS             string `json:"naics"`
	Type              string `json:"type"`
	PostedDate        string `json:"posted_date"`
	ResponseDeadline  string `json:"response_deadline"`
	AwardDate         string `json:"award_date"`
	ContractingOffice string `json:"contracting_office"`
	Agency            string `json:"agency"`
}

func main() {
	db.InitDB()
	cronjob.StartScheduler()
	openai.Init()

	// GET /contracts — list all active contracts
	http.HandleFunc("/contracts", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.DB.Query(`
			SELECT id, title, description, status, naics, type, posted_date,
			       response_deadline, award_date, contracting_office, agency
			FROM contracts
			WHERE status = 'active'
		`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var contracts []Contract
		for rows.Next() {
			var c Contract
			if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &c.NAICS,
				&c.Type, &c.PostedDate, &c.ResponseDeadline, &c.AwardDate,
				&c.ContractingOffice, &c.Agency); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contracts = append(contracts, c)
		}

		w.Header().Set("Content-Type", "application/json")
		output, err := json.MarshalIndent(contracts, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(output)
	})

	// GET /contracts/naics/{code}
	http.HandleFunc("/contracts/naics/", func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/contracts/naics/")
		if code == "" {
			http.Error(w, "NAICS code required", http.StatusBadRequest)
			return
		}

		rows, err := db.DB.Query(`
			SELECT id, title, description, status, naics, type, posted_date,
			       response_deadline, award_date, contracting_office, agency
			FROM contracts
			WHERE status = 'active' AND TRIM(naics) = ?
		`, code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var contracts []Contract
		for rows.Next() {
			var c Contract
			if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &c.NAICS,
				&c.Type, &c.PostedDate, &c.ResponseDeadline, &c.AwardDate,
				&c.ContractingOffice, &c.Agency); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contracts = append(contracts, c)
		}

		w.Header().Set("Content-Type", "application/json")
		output, err := json.MarshalIndent(contracts, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(output)
	})

	// GET /contracts/search?query=...
	http.HandleFunc("/contracts/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		if query == "" {
			http.Error(w, "`query` parameter required", http.StatusBadRequest)
			return
		}

		rows, err := db.DB.Query("SELECT id, title FROM contracts WHERE status='active'")
		if err != nil {
			http.Error(w, "DB query failed", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var summaries []string
		idToFull := make(map[string]Contract)

		for rows.Next() {
			var id, title, desc string
			if err := rows.Scan(&id, &title, &desc); err != nil {
				continue
			}
			summary := id + ": " + title + " — " + desc
			summaries = append(summaries, summary)

			// preload full row
			row := db.DB.QueryRow(`
				SELECT id, title, description, status, naics, type,
				       posted_date, response_deadline, award_date,
				       contracting_office, agency
				FROM contracts WHERE id = ?`, id)

			var c Contract
			if err := row.Scan(&c.ID, &c.Title, &c.Description, &c.Status, &c.NAICS,
				&c.Type, &c.PostedDate, &c.ResponseDeadline, &c.AwardDate,
				&c.ContractingOffice, &c.Agency); err == nil {
				idToFull[id] = c
			}
		}

		topIDs, err := openai.ScoreContracts(query, summaries)
		if err != nil {
			http.Error(w, "ChatGPT error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var result []Contract
		for _, id := range topIDs {
			if c, ok := idToFull[id]; ok {
				result = append(result, c)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		output, err := json.MarshalIndent(contracts, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(output)
	})

	log.Println("🚀 Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
