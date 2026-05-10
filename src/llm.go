package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AnthropicRequest struct {
	Model    string              `json:"model"`
	System   string              `json:"system"`
	Messages []AnthropicMessage  `json:"messages"`
	MaxTokens int                `json:"max_tokens"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

type LLMDecision struct {
	Decision      string `json:"decision"`
	Reasoning     string `json:"reasoning"`
	ReasoningLong string `json:"reasoning_long"`
}

func callLLM(cfg *LLMConfig, systemPrompt, userPrompt string) (*LLMDecision, error) {
	url := strings.TrimRight(cfg.BaseURL, "/") + "/messages"
	timeout := cfg.ResolveTimeout()

	body := AnthropicRequest{
		Model:    cfg.Model,
		System:   systemPrompt,
		Messages: []AnthropicMessage{
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 512,
	}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("POST", url, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm returned %d", resp.StatusCode)
	}

	var llmResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(llmResp.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	content := llmResp.Content[0].Text
	content = extractJSON(content)

	var dec LLMDecision
	if err := json.Unmarshal([]byte(content), &dec); err != nil {
		return nil, fmt.Errorf("parse decision: %w (got: %s)", err, content[:min(200, len(content))])
	}
	return &dec, nil
}

func extractJSON(s string) string {
	start := 0
	end := len(s)
	for i, c := range s {
		if c == '{' {
			start = i
			break
		}
	}
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '}' {
			end = i + 1
			break
		}
	}
	if start < end {
		return s[start:end]
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
