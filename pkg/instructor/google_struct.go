package instructor

import (
	"google.golang.org/genai"
)

type InstructorGoogle struct {
	*genai.Client

	provider   Provider
	mode       Mode
	maxRetries int
	validate   bool
}

func FromGoogle(client *genai.Client, opts ...Options) *InstructorGoogle {
	options := mergeOptions(opts...)

	i := &InstructorGoogle{
		Client: client,

		provider:   ProviderGoogle,
		mode:       *options.Mode,
		maxRetries: *options.MaxRetries,
		validate:   *options.validate,
	}
	return i
}

func (i *InstructorGoogle) Provider() Provider {
	return i.provider
}

func (i *InstructorGoogle) Mode() Mode {
	return i.mode
}

func (i *InstructorGoogle) MaxRetries() int {
	return i.maxRetries
}

func (i *InstructorGoogle) Validate() bool {
	return i.validate
}

// GoogleRequest represents a request to the Google AI API
type GoogleRequest struct {
	Model            string                  `json:"model"`
	Contents         []*genai.Content        `json:"contents"`
	GenerationConfig *genai.GenerationConfig `json:"generationConfig,omitempty"`
	SafetySettings   []*genai.SafetySetting  `json:"safetySettings,omitempty"`
}

// GoogleResponse represents a response from the Google AI API
type GoogleResponse struct {
	Candidates    []*genai.Candidate                          `json:"candidates"`
	UsageMetadata *genai.GenerateContentResponseUsageMetadata `json:"usageMetadata,omitempty"`
}

func (i *InstructorGoogle) emptyResponseWithUsageSum(usage *UsageSum) interface{} {
	return &GoogleResponse{
		UsageMetadata: &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(usage.InputTokens),
			CandidatesTokenCount: int32(usage.OutputTokens),
			TotalTokenCount:      int32(usage.TotalTokens),
		},
	}
}

func (i *InstructorGoogle) emptyResponseWithResponseUsage(response interface{}) interface{} {
	resp, ok := response.(*GoogleResponse)
	if !ok || resp == nil {
		return nil
	}

	return &GoogleResponse{
		UsageMetadata: resp.UsageMetadata,
	}
}

func (i *InstructorGoogle) addUsageSumToResponse(response interface{}, usage *UsageSum) (interface{}, error) {
	resp, ok := response.(*GoogleResponse)
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

func (i *InstructorGoogle) countUsageFromResponse(response interface{}, usage *UsageSum) *UsageSum {
	resp, ok := response.(*GoogleResponse)
	if !ok || resp == nil || resp.UsageMetadata == nil {
		return usage
	}

	usage.InputTokens = int(resp.UsageMetadata.PromptTokenCount)
	usage.OutputTokens = int(resp.UsageMetadata.CandidatesTokenCount)
	usage.TotalTokens = int(resp.UsageMetadata.TotalTokenCount)

	return usage
}

func nilGoogleRespWithUsage(resp *GoogleResponse) *GoogleResponse {
	if resp == nil {
		return &GoogleResponse{}
	}
	return resp
}
