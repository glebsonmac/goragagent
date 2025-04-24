package integration

import (
	"os"
	"testing"

	"goragagent/cmd"

	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestOpenAIIntegration(t *testing.T) {
	// Skip if no API key is set
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping OpenAI integration test: OPENAI_API_KEY not set")
	}

	client := openai.NewClient(apiKey)
	contextInfo := "According to tax_policies_2023.pdf, the tax rate in Travis County is 1.9%"
	followUp := ""
	question := "What's the tax rate in Travis County?"

	answer, err := cmd.GenerateAnswer(client, contextInfo, followUp, question)
	assert.NoError(t, err)
	assert.NotEmpty(t, answer)
	assert.Contains(t, answer, "1.9%")
}
