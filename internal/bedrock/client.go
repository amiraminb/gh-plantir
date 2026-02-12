package bedrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/amiraminb/gh-plantir/internal/github"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

type Client struct {
	runtime *bedrockruntime.Client
	model   string
}

func NewClient(profile, region, model string) (*Client, error) {
	var opts []func(*awsconfig.LoadOptions) error
	opts = append(opts, awsconfig.WithRegion(region))
	if profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		runtime: bedrockruntime.NewFromConfig(cfg),
		model:   model,
	}, nil
}

type bedrockRequest struct {
	AnthropicVersion string           `json:"anthropic_version"`
	MaxTokens        int              `json:"max_tokens"`
	System           string           `json:"system"`
	Messages         []bedrockMessage `json:"messages"`
}

type bedrockMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bedrockResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) Summarize(prs []github.MergedPR) (string, error) {
	prompt, err := BuildPrompt(prs)
	if err != nil {
		return "", err
	}

	reqBody := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		System:           systemPrompt,
		Messages: []bedrockMessage{
			{Role: "user", Content: prompt},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	out, err := c.runtime.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.model),
		ContentType: aws.String("application/json"),
		Body:        payload,
	})
	if err != nil {
		return "", fmt.Errorf("failed to invoke Bedrock model: %w", err)
	}

	var resp bedrockResponse
	if err := json.Unmarshal(out.Body, &resp); err != nil {
		return "", fmt.Errorf("failed to parse Bedrock response: %w", err)
	}

	if len(resp.Content) == 0 {
		return "", fmt.Errorf("empty response from Bedrock")
	}

	return resp.Content[0].Text, nil
}
