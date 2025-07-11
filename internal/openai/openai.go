package openai

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func Init() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client = openai.NewClient(apiKey)
}

// ScoreContracts sends the query and contract summaries to OpenAI,
// and expects a comma-separated list of ranked contract IDs.
func ScoreContracts(query string, contracts []string) ([]string, error) {
	// Build a clearer prompt
	prompt := fmt.Sprintf(`
The search term is: "%s".

Below is a list of contract summaries. Each summary is formatted as "ID: title — description":

%s

Rank the contracts by how relevant they are to the search term. 
ONLY return a comma-separated list of the contract IDs in order of relevance. 
Do not include any additional explanation, formatting, or text.
`, query, strings.Join(contracts, "\n"))

	log.Printf("📤 OpenAI request prompt:\n%s", prompt)

	// Send to OpenAI
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a helpful assistant that ranks government contract opportunities based on relevance to a given search term."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens: 100,
		},
	)
	log.Printf("*** OpenAI raw response: %+v", resp)
	if err != nil {
		return nil, err
	}

	// Extract and log response
	content := resp.Choices[0].Message.Content
	log.Printf("✅ OpenAI message content: %s", content)

	// Parse comma-separated contract IDs
	rawIDs := strings.Split(content, ",")
	var results []string
	for _, id := range rawIDs {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			results = append(results, trimmed)
		}
	}

	return results, nil
}
