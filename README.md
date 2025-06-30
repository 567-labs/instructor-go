# instructor-go - Structured LLM Outputs

Instructor Go is a library that makes it a breeze to work with structured outputs from large language models (LLMs).

---

[![Twitter Follow](https://img.shields.io/twitter/follow/jxnlco?style=social)](https://twitter.com/jxnlco)
[![LinkedIn Follow](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/robby-horvath/)
[![Documentation](https://img.shields.io/badge/docs-available-brightgreen)](https://go.useinstructor.com)
[![GitHub issues](https://img.shields.io/github/issues/instructor-ai/instructor-go.svg)](https://github.com/instructor-ai/instructor-go/issues)
[![Discord](https://img.shields.io/discord/1192334452110659664?label=discord)](https://discord.gg/UD9GPjbs8c)

Built on top of [`invopop/jsonschema`](https://github.com/invopop/jsonschema) and utilizing `jsonschema` Go struct tags (so no changing code logic), it provides a simple and user-friendly API to manage validation, retries, and streaming responses. Get ready to supercharge your LLM workflows!

## Install

Install the package into your code with:

```bash
go get "github.com/instructor-ai/instructor-go/pkg/instructor"
```

Import in your code:

```go
import (
	"github.com/instructor-ai/instructor-go/pkg/instructor"
)
```

## Example

As shown in the example below, by adding extra metadata to each struct field (via `jsonschema` tag) we want the model to be made aware of:

> For more information on the `jsonschema` tags available, see the [`jsonschema` godoc](https://pkg.go.dev/github.com/invopop/jsonschema?utm_source=godoc).

Running

```bash
export OPENAI_API_KEY=<Your OpenAI API Key>
go run examples/user/main.go
```

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/instructor-ai/instructor-go/pkg/instructor"
	openai "github.com/sashabaranov/go-openai"
)

type Person struct {
	Name string `json:"name"          jsonschema:"title=the name,description=The name of the person,example=joe,example=lucy"`
	Age  int    `json:"age,omitempty" jsonschema:"title=the age,description=The age of the person,example=25,example=67"`
}

func main() {
	ctx := context.Background()

	client := instructor.FromOpenAI(
		openai.NewClient(os.Getenv("OPENAI_API_KEY")),
		instructor.WithMode(instructor.ModeJSON),
		instructor.WithMaxRetries(3),
	)

	var person Person
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Extract Robby is 22 years old.",
				},
			},
		},
		&person,
	)
	_ = resp // sends back original response so no information loss from original API
	if err != nil {
		panic(err)
	}

	fmt.Printf(`
Name: %s
Age:  %d
`, person.Name, person.Age)
	/*
		Name: Robby
		Age:  22
	*/
}
```

See all examples here [`examples/README.md`](examples/README.md)

## Providers

Instructor Go supports the following LLM provider APIs:
- [OpenAI](https://github.com/sashabaranov/go-openai)
- [Anthropic](https://github.com/liushuangls/go-anthropic)
- [Cohere](github.com/cohere-ai/cohere-go)
- [Google](github.com/googleapis/go-genai)

### Usage (token counts)

These provider APIs include usage data (input and output token counts) in their responses, which Instructor Go captures and returns in the response object.

Usage is summed for retries. If multiple requests are needed to get a valid response, the usage from all requests is summed and returned. Even if Instructor fails to get a valid response after the maximum number of retries, the usage sum from all attempts is still returned.

### How to view usage data

<details>
<summary>Usage counting with OpenAI</summary>

```go
resp, err := client.CreateChatCompletion(
    ctx,
    openai.ChatCompletionRequest{
        Model: openai.GPT4o,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Extract Robby is 22 years old.",
            },
        },
    },
    &person,
)

fmt.Printf("Input tokens: %d\n", resp.Usage.PromptTokens)
fmt.Printf("Output tokens: %d\n", resp.Usage.CompletionTokens)
fmt.Printf("Total tokens: %d\n", resp.Usage.TotalTokens)
```

</details>

<details>
<summary>Usage counting with Anthropic</summary>

```go
resp, err := client.CreateMessages(
    ctx,
    anthropic.MessagesRequest{
        Model: anthropic.ModelClaude3Haiku20240307,
        Messages: []anthropic.Message{
            anthropic.NewUserTextMessage("Classify the following support ticket: My account is locked and I can't access my billing info."),
        },
		MaxTokens: 500,
    },
    &prediction,
)

fmt.Printf("Input tokens: %d\n", resp.Usage.InputTokens)
fmt.Printf("Output tokens: %d\n", resp.Usage.OutputTokens)
```

</details>

<details>
<summary>Usage counting with Cohere</summary>

```go
resp, err := client.Chat(
    ctx,
    &cohere.ChatRequest{
        Model: "command-r-plus",
        Message: "Tell me about the history of artificial intelligence up to year 2000",
        MaxTokens: 2500,
    },
    &historicalFact,
)

fmt.Printf("Input tokens: %d\n", int(*resp.Meta.Tokens.InputTokens))
fmt.Printf("Output tokens: %d\n", int(*resp.Meta.Tokens.OutputTokens))
```

</details>