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
		MaxTokens: 1024,
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

	var rawResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var textContent string
	if contentBlocks, ok := rawResp["content"].([]interface{}); ok {
		for _, block := range contentBlocks {
			if b, ok := block.(map[string]interface{}); ok {
				if txt, _ := b["text"].(string); txt != "" {
					textContent = txt
					break
				}
			}
		}
	}

	// Text block not found — some thinking models only produce thinking blocks.
	// Fall back to extracting JSON from the thinking content.
	if textContent == "" {
		if contentBlocks, ok := rawResp["content"].([]interface{}); ok {
			for _, block := range contentBlocks {
				if b, ok := block.(map[string]interface{}); ok {
					if think, _ := b["thinking"].(string); think != "" {
						textContent = extractJSON(think)
						if textContent != "" {
							break
						}
					}
				}
			}
		}
	}
	if textContent == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	textContent = extractJSON(textContent)

	var dec LLMDecision
	if err := json.Unmarshal([]byte(textContent), &dec); err != nil {
		return nil, fmt.Errorf("parse decision: %w (got: %s)", err, textContent[:min(200, len(textContent))])
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
