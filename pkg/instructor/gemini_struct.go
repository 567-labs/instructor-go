package instructor

import (
	"google.golang.org/genai"
)

type InstructorGemini struct {
	*genai.Client

	provider   Provider
	mode       Mode
	maxRetries int
	validate   bool
}

func FromGemini(client *genai.Client, opts ...Options) *InstructorGemini {
	options := mergeOptions(opts...)

	i := &InstructorGemini{
		Client: client,

		provider:   ProviderGemini,
		mode:       *options.Mode,
		maxRetries: *options.MaxRetries,
		validate:   *options.validate,
	}
	return i
}

func (i *InstructorGemini) Provider() Provider {
	return i.provider
}

func (i *InstructorGemini) Mode() Mode {
	return i.mode
}

func (i *InstructorGemini) MaxRetries() int {
	return i.maxRetries
}

func (i *InstructorGemini) Validate() bool {
	return i.validate
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Model            string                  `json:"model"`
	Contents         []*genai.Content        `json:"contents"`
	GenerationConfig *genai.GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []*genai.SafetySetting  `json:"safetySettings,omitempty"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates    []*genai.Candidate                          `json:"candidates"`
	UsageMetadata *genai.GenerateContentResponseUsageMetadata `json:"usageMetadata,omitempty"`
}

func (i *InstructorGemini) emptyResponseWithUsageSum(usage *UsageSum) interface{} {
	return &GeminiResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(usage.InputTokens),
			CandidatesTokenCount: int32(usage.OutputTokens),
			TotalTokenCount:      int32(usage.TotalTokens),
		},
	}
}

func (i *InstructorGemini) emptyResponseWithResponseUsage(response interface{}) interface{} {
	resp, ok := response.(*GeminiResponse)
	if !ok || resp == nil {
		return nil
	}

	return &GeminiResponse{
		UsageMetadata: resp.UsageMetadata,
	}
}

func (i *InstructorGemini) addUsageSumToResponse(response interface{}, usage *UsageSum) (interface{}, error) {
	resp, ok := response.(*GeminiResponse)
	if !ok || resp == nil {
		return response, nil
	}

	if resp.UsageMetadata == nil {
		resp.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{}
	}

	resp.UsageMetadata.PromptTokenCount = int32(usage.InputTokens)
	resp.UsageMetadata.CandidatesTokenCount = int32(usage.OutputTokens)
	resp.UsageMetadata.TotalTokenCount = int32(usage.TotalTokens)

	return resp, nil
}

func (i *InstructorGemini) countUsageFromResponse(response interface{}, usage *UsageSum) *UsageSum {
	resp, ok := response.(*GeminiResponse)
	if !ok || resp == nil || resp.UsageMetadata == nil {
		return usage
	}

	usage.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
	usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	usage.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)

	return usage
}

func nilGeminiRespWithUsage(resp *GeminiResponse) *GeminiResponse {
	if resp == nil {
		return &GeminiResponse{}
	}
	return resp
}
