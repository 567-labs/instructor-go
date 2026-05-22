package openai

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestFirstChoiceMessageContent(t *testing.T) {
	t.Run("empty choices returns error", func(t *testing.T) {
		_, err := firstChoiceMessageContent(&openai.ChatCompletionResponse{})
		if err == nil {
			t.Fatal("expected error for empty choices")
		}
	})

	t.Run("returns first choice content", func(t *testing.T) {
		content, err := firstChoiceMessageContent(&openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: `{"ok":true}`,
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if content != `{"ok":true}` {
			t.Fatalf("content = %q, want %q", content, `{"ok":true}`)
		}
	})
}
