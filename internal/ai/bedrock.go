package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const (
	DefaultModel   = "us.anthropic.claude-opus-4-5-20251101-v1:0"
	DefaultRegion  = "us-west-2"
	DefaultProfile = "tooling-ai-coding-assistant"
)

type Client struct {
	bedrock *bedrockruntime.Client
	model   string
	profile string
	region  string
}

type ReviewResult struct {
	Summary    string
	CodeReview string
}

type PRContext struct {
	Title       string
	URL         string
	Description string
	Author      string
	Labels      []string
	Additions   int
	Deletions   int
	Files       []string
	BaseBranch  string
	HeadBranch  string
	Diff        string
}

func NewClient(profile, region, model string) (*Client, error) {
	if profile == "" {
		profile = DefaultProfile
	}
	if region == "" {
		region = DefaultRegion
	}
	if model == "" {
		model = DefaultModel
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	return &Client{
		bedrock: client,
		model:   model,
		profile: profile,
		region:  region,
	}, nil
}

func (c *Client) refreshCredentials() error {
	fmt.Println("AWS session expired. Logging in...")
	cmd := exec.Command("aws", "sso", "login", "--profile", c.profile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSO login failed: %w", err)
	}

	// Recreate the client with fresh credentials
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(c.profile),
		config.WithRegion(c.region),
	)
	if err != nil {
		return fmt.Errorf("failed to reload AWS config: %w", err)
	}

	c.bedrock = bedrockruntime.NewFromConfig(cfg)
	return nil
}

func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "ExpiredToken") ||
		strings.Contains(errStr, "InvalidIdentityToken") ||
		strings.Contains(errStr, "UnauthorizedException") ||
		strings.Contains(errStr, "AccessDeniedException") ||
		strings.Contains(errStr, "security token") ||
		strings.Contains(errStr, "credentials")
}

type claudeRequest struct {
	AnthropicVersion string    `json:"anthropic_version"`
	MaxTokens        int       `json:"max_tokens"`
	Messages         []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func (c *Client) ReviewPR(ctx PRContext) (*ReviewResult, error) {
	// Build file list
	fileList := ""
	if len(ctx.Files) > 0 {
		fileList = "\n\nFiles changed:\n"
		for _, f := range ctx.Files {
			fileList += fmt.Sprintf("- %s\n", f)
		}
	}

	// Build labels
	labelStr := ""
	if len(ctx.Labels) > 0 {
		labelStr = fmt.Sprintf("\nLabels: %s", strings.Join(ctx.Labels, ", "))
	}

	// Build description section
	descSection := ""
	if ctx.Description != "" {
		descSection = fmt.Sprintf("\n\n## PR Description\n%s", ctx.Description)
	}

	prompt := fmt.Sprintf(`You are a senior software engineer reviewing a pull request.

## PR Information
- **Title:** %s
- **Author:** %s
- **Branch:** %s → %s
- **Changes:** +%d -%d lines%s%s
- **URL:** %s%s

## Diff

%s

---

Please provide:

## Summary
A brief summary (2-4 sentences) of what this PR does. Focus on the intent and main changes.

## Code Review
Review the code changes and provide feedback on:
- Potential bugs or issues
- Security concerns (injection, auth, data exposure, etc.)
- Performance considerations
- Code quality and best practices
- Error handling
- Suggestions for improvement

Be concise and actionable. Focus on important issues. If the code looks good, say so briefly.`,
		ctx.Title,
		ctx.Author,
		ctx.HeadBranch,
		ctx.BaseBranch,
		ctx.Additions,
		ctx.Deletions,
		labelStr,
		fileList,
		ctx.URL,
		descSection,
		ctx.Diff)

	request := claudeRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	output, err := c.invokeWithRetry(body)
	if err != nil {
		return nil, err
	}

	var response claudeResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from model")
	}

	return &ReviewResult{
		Summary:    response.Content[0].Text,
		CodeReview: "", // The response contains both in one text block
	}, nil
}

func (c *Client) invokeWithRetry(body []byte) (*bedrockruntime.InvokeModelOutput, error) {
	output, err := c.bedrock.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.model),
		ContentType: aws.String("application/json"),
		Body:        body,
	})

	if err != nil && isAuthError(err) {
		// Try to refresh credentials and retry
		if refreshErr := c.refreshCredentials(); refreshErr != nil {
			return nil, fmt.Errorf("failed to invoke model: %w (and SSO login failed: %v)", err, refreshErr)
		}

		// Retry the request
		output, err = c.bedrock.InvokeModel(context.Background(), &bedrockruntime.InvokeModelInput{
			ModelId:     aws.String(c.model),
			ContentType: aws.String("application/json"),
			Body:        body,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to invoke model: %w", err)
	}

	return output, nil
}
