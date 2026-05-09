package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ---- LLM types ----

type LLMRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type LLMDecision struct {
	Decision       string `json:"decision"`
	Reasoning      string `json:"reasoning"`
	ReasoningLong  string `json:"reasoning_long"`
}

// ---- LLM call ----

func callLLM(cfg *LLMConfig, systemPrompt, userPrompt string) (*LLMDecision, error) {
	url := cfg.ResolveBaseURL()
	timeout := cfg.ResolveTimeout()

	body := LLMRequest{
		Model: cfg.Model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
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
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("llm returned %d", resp.StatusCode)
	}

	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	content := llmResp.Choices[0].Message.Content
	// extract JSON from response (may have markdown wrapping)
	content = extractJSON(content)

	var dec LLMDecision
	if err := json.Unmarshal([]byte(content), &dec); err != nil {
		return nil, fmt.Errorf("parse decision: %w (got: %s)", err, content[:200])
	}
	return &dec, nil
}

func extractJSON(s string) string {
	// find first { and last }
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
