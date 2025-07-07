# GovRFP Finder

**GovRFP Finder** is a lightweight and scalable Go application that enables users to search, filter, and discover government RFP (Request for Proposal) contracts from public procurement sources. It is designed for contractors, agencies, and businesses seeking to automate the process of identifying public contract opportunities.

## Features
- Search and filter active RFP contracts by:
    - Keywords
    - NAICS codes
    - Deadlines
- Built with Go, prioritizing concurrency and speed
- Utilizes OpenAI for contract ranking based on relevance
- Offers a RESTful API for seamless integration with applications or CLI tools
- Conducts periodic cleanup of expired RFPs
- Optional scheduled cron jobs for task automation

## Project Structure
```
go-rfp/
├── cmd/api/          # Main HTTP API server
├── internal/db/      # Database connection and queries
├── internal/cronjob/ # Scheduled cleanup tasks
├── internal/openai/  # OpenAI relevance scoring integration
├── go.mod
└── README.md
```

## Requirements
- Go 1.20+
- PostgreSQL or SQLite
- OpenAI API Key (OPENAI_API_KEY environment variable)

## Setup & Run
1. Clone the repo
   ```
   git clone https://github.com/jackmbuda/go-rfp.git
   cd go-rfp
   ```

2. Install dependencies
   ```
   go mod tidy
   ```

3. Set environment variables
   ```
   export DATABASE_URL="postgres://username:password@localhost:5432/rfp_db?sslmode=disable"
   export OPENAI_API_KEY="sk-..."
   ```
   (For SQLite usage, configure it within internal/db)

4. Start the server
   ```
   go run ./cmd/api
   ```
   The server will run at http://localhost:8080.

## API Endpoints
| Method | Endpoint                  | Description                          |
| ------ | ------------------------- | ------------------------------------ |
| GET    | /contracts                | List all active contracts            |
| GET    | /contracts/naics/{naics_code} | List contracts by NAICS code      |
| GET    | /contracts/search?query=...   | Search contracts using OpenAI relevance |

## Example API Calls
- `curl http://localhost:8080/contracts`
- `curl http://localhost:8080/contracts/naics/541512`
- `curl "http://localhost:8080/contracts/search?query=cybersecurity"`

## Cleanup
Expired contracts (where response_deadline < today) are automatically removed upon startup and daily through a cron job.

## License
MIT License.

## Contributing
PRs, issues, and feature requests are welcome! Feel free to fork the repo and submit your contributions.

## Acknowledgments
- US Government public procurement portals
- OpenAI API for smart contract scoring
- Go community for their excellent project structures
