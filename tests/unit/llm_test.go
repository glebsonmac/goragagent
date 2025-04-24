package unit

import (
	"testing"

	"goragagent/cmd"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAnswerWithNilClient(t *testing.T) {
	t.Log("Testing LLM response generation with nil client (fallback mode)...")

	contextInfo := "Test context"
	question := "Test question"

	t.Logf("Context: %s", contextInfo)
	t.Logf("Question: %s", question)

	// Test with nil client
	answer, err := cmd.GenerateAnswer(nil, contextInfo, "", question)
	assert.NoError(t, err, "Should not error with nil client")
	assert.Equal(t, contextInfo, answer, "Should return context as answer in fallback mode")
	t.Log("✓ Successfully handled nil client case")
}

func TestGenerateAnswerBasicResponse(t *testing.T) {
	t.Log("Testing LLM basic response with real data...")

	contextInfo := "According to tax_policies_2023.pdf, the tax rate in Travis County is 1.9%"
	question := "What's the tax rate in Travis County?"

	t.Logf("Context: %s", contextInfo)
	t.Logf("Question: %s", question)

	// Test without OpenAI client (basic response mode)
	answer, err := cmd.GenerateAnswer(nil, contextInfo, "", question)
	assert.NoError(t, err, "Should not error in basic response mode")
	assert.Equal(t, contextInfo, answer, "Should return context as answer in basic mode")
	t.Log("✓ Successfully generated basic response")
}
