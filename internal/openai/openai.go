package openai

import (
	"context"
	"fmt"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func Init() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	client = openai.NewClient(apiKey)
}

// scoreContracts sends the query and returns the top 5 contract IDs
func ScoreContracts(query string, contracts []string) ([]string, error) {
	prompt := fmt.Sprintf(`Given this search query: %q
Rank these contract IDs in order of relevance from most to least:
%s`, query, "```\n"+strings.Join(contracts, "\n")+"```")

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a contract finder."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens: 150,
		},
	)
	if err != nil {
		return nil, err
	}

	content := resp.Choices[0].Message.Content
	lines := strings.Split(content, "\n")
	var results []string
	for _, line := range lines {
		id := strings.TrimSpace(line)
		if id != "" {
			results = append(results, id)
		}
	}
	return results, nil
}
