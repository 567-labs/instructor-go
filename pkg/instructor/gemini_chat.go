package instructor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

// No GeminiResponse struct here; use the one from gemini_struct.go

func (i *InstructorGemini) CreateChatCompletion(
	ctx context.Context,
	request GeminiRequest,
	responseType any,
) (response GeminiResponse, err error) {
	resp, err := chatHandler(i, ctx, request, responseType)
	if err != nil {
		if resp == nil {
			return GeminiResponse{}, err
		}
		return *nilGeminiRespWithUsage(resp.(*GeminiResponse)), err
	}
	response = *(resp.(*GeminiResponse))
	return response, nil
}

func (i *InstructorGemini) chat(ctx context.Context, request interface{}, schema *Schema) (string, interface{}, error) {
	req, ok := request.(GeminiRequest)
	if !ok {
		return "", nil, fmt.Errorf("invalid request type for %s client", i.Provider())
	}
	switch i.Mode() {
	case ModeToolCall:
		return i.chatToolCall(ctx, &req, schema, false)
	case ModeToolCallStrict:
		return i.chatToolCall(ctx, &req, schema, true)
	case ModeJSON:
		return i.chatJSON(ctx, &req, schema, false)
	case ModeJSONStrict:
		return i.chatJSON(ctx, &req, schema, true)
	case ModeJSONSchema:
		return i.chatJSONSchema(ctx, &req, schema)
	default:
		return "", nil, fmt.Errorf("mode '%s' is not supported for %s", i.Mode(), i.Provider())
	}
}

func (i *InstructorGemini) chatToolCall(ctx context.Context, request *GeminiRequest, schema *Schema, strict bool) (string, *GeminiResponse, error) {
	tools := createGeminiTools(schema, strict)
	if request.GenerationConfig == nil {
		request.GenerationConfig = &genai.GenerationConfig{}
	}
	resp, err := i.Models.GenerateContent(ctx, request.Model, request.Contents, &genai.GenerateContentConfig{
		Tools:          tools,
		SafetySettings: request.SafetySettings,
	})
	if err != nil {
		return "", nil, err
	}
	geminiResp := &GeminiResponse{
		Candidates:    resp.Candidates,
		UsageMetadata: resp.UsageMetadata,
	}
	var toolCalls []*genai.FunctionCall
	for _, candidate := range resp.Candidates {
		if candidate.Content != nil {
			for _, part := range candidate.Content.Parts {
				if part.FunctionCall != nil {
					toolCalls = append(toolCalls, part.FunctionCall)
				}
			}
		}
	}
	numTools := len(toolCalls)
	if numTools < 1 {
		return "", nilGeminiRespWithUsage(geminiResp), errors.New("received no tool calls from model, expected at least 1")
	}
	if numTools == 1 {
		argsJSON, err := json.Marshal(toolCalls[0].Args)
		if err != nil {
			return "", nilGeminiRespWithUsage(geminiResp), err
		}
		return string(argsJSON), geminiResp, nil
	}
	jsonArray := make([]map[string]interface{}, len(toolCalls))
	for i, toolCall := range toolCalls {
		argsJSON, err := json.Marshal(toolCall.Args)
		if err != nil {
			return "", nilGeminiRespWithUsage(geminiResp), err
		}
		var jsonObj map[string]interface{}
		err = json.Unmarshal(argsJSON, &jsonObj)
		if err != nil {
			return "", nilGeminiRespWithUsage(geminiResp), err
		}
		jsonArray[i] = jsonObj
	}
	resultJSON, err := json.Marshal(jsonArray)
	if err != nil {
		return "", nilGeminiRespWithUsage(geminiResp), err
	}
	return string(resultJSON), geminiResp, nil
}

func (i *InstructorGemini) chatJSON(ctx context.Context, request *GeminiRequest, schema *Schema, strict bool) (string, *GeminiResponse, error) {
	structName := schema.NameFromRef()
	request.Contents = prependGeminiContents(request.Contents, *createGeminiJSONMessage(schema))
	if strict {
		if request.GenerationConfig == nil {
			request.GenerationConfig = &genai.GenerationConfig{}
		}
	}
	resp, err := i.Models.GenerateContent(ctx, request.Model, request.Contents, &genai.GenerateContentConfig{
		SafetySettings: request.SafetySettings,
	})
	if err != nil {
		return "", nil, err
	}
	geminiResp := &GeminiResponse{
		Candidates:    resp.Candidates,
		UsageMetadata: resp.UsageMetadata,
	}
	text := ""
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				text = part.Text
				break
			}
		}
	}
	if strict {
		resMap := make(map[string]any)
		_ = json.Unmarshal([]byte(text), &resMap)
		cleanedText, _ := json.Marshal(resMap[structName])
		text = string(cleanedText)
	}
	return text, geminiResp, nil
}

func (i *InstructorGemini) chatJSONSchema(ctx context.Context, request *GeminiRequest, schema *Schema) (string, *GeminiResponse, error) {
	request.Contents = prependGeminiContents(request.Contents, *createGeminiJSONMessage(schema))
	resp, err := i.Models.GenerateContent(ctx, request.Model, request.Contents, &genai.GenerateContentConfig{
		SafetySettings: request.SafetySettings,
	})
	if err != nil {
		return "", nil, err
	}
	geminiResp := &GeminiResponse{
		Candidates:    resp.Candidates,
		UsageMetadata: resp.UsageMetadata,
	}
	text := ""
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				text = part.Text
				break
			}
		}
	}
	return text, geminiResp, nil
}

func createGeminiJSONMessage(schema *Schema) *genai.Content {
	schemaJSON, _ := json.Marshal(schema.Schema)
	return &genai.Content{
		Parts: []*genai.Part{
			{
				Text: fmt.Sprintf("You are a helpful assistant that responds with valid JSON according to the following schema:\n\n%s\n\nRespond with valid JSON only.", string(schemaJSON)),
			},
		},
		Role: "user",
	}
}

func createGeminiTools(schema *Schema, strict bool) []*genai.Tool {
	// TODO: Convert schema.Schema.Properties to map[string]*genai.Schema if needed
	tool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        schema.NameFromRef(),
				Description: schema.Description,
				Parameters: &genai.Schema{
					Type:       "object",
					Properties: map[string]*genai.Schema{}, // TODO: convert from schema.Schema.Properties
					Required:   []string{},                 // TODO: convert from schema.Schema.Required
				},
			},
		},
	}
	return []*genai.Tool{tool}
}

func prependGeminiContents(contents []*genai.Content, content genai.Content) []*genai.Content {
	return append([]*genai.Content{&content}, contents...)
}
