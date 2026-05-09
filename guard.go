package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ---- output types ----

const (
	DecisionAllow = "allow"
	DecisionAsk   = "ask"
)

type Decision struct {
	Decision              string `json:"decision"`
	Reason                string `json:"reason"`
	SystemMessage         string `json:"systemMessage,omitempty"`
	HookSpecificOutput    *HookOutput `json:"hookSpecificOutput"`
}

type HookOutput struct {
	HookEventName             string `json:"hookEventName"`
	PermissionDecision        string `json:"permissionDecision"`
	PermissionDecisionReason  string `json:"permissionDecisionReason,omitempty"`
}

// ---- transcript reading ----

func readUserMessages(transcriptPath string) string {
	if transcriptPath == "" {
		return ""
	}
	data, err := os.ReadFile(transcriptPath)
	if err != nil {
		return ""
	}
	text := string(data)
	// Take tail for context window
	if len(text) > 4000 {
		text = text[len(text)-4000:]
	}
	// Extract user messages only (reasoning-blind)
	// Simple heuristic: lines with "user:" prefix or user role
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Keep lines with clear user attribution
		if strings.Contains(trimmed, `"role":"user"`) ||
			strings.Contains(trimmed, `"role": "user"`) ||
			strings.HasPrefix(trimmed, "user:") ||
			strings.HasPrefix(trimmed, "User:") {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) == 0 {
		// Fallback: take raw tail
		if len(text) > 2000 {
			text = text[len(text)-2000:]
		}
		return text
	}
	return strings.Join(lines, "\n")
}

// ---- main classification logic ----

func classify(req *HookRequest) *Decision {
	toolName := req.CanonicalTool()

	// Tier 1: always-allow tools
	if tier1Tools[toolName] {
		return allowDecision("tier 1: safe read-only tool")
	}

	// Tier 1: Bash with read-only patterns
	if toolName == "Bash" {
		if isReadOnlyBash(req.ToolInput.Command()) {
			return allowDecision("tier 1: read-only bash command")
		}
		// Tier 2: Bash safe writes inside project
		if isTier2Bash(req.ToolInput.Command()) {
			target := extractBashTarget(req.ToolInput.Command())
			if target != "" && isInsideProject(target) {
				return allowDecision("tier 2: safe bash write inside project")
			}
		}
	}

	// Tier 2: Write/Edit inside project
	if tier2Tools[toolName] {
		filePath := req.ToolInput.FilePath()
		if filePath != "" && isInsideProject(filePath) {
			return allowDecision("tier 2: safe edit inside project")
		}
	}

	// Tier 3: LLM review
	cfg, err := LoadConfig()
	if err != nil {
		log.Printf("auto-guard: config error: %v", err)
		return askDecision("config error, asking user")
	}

	userMsgs := readUserMessages(req.TranscriptPath)
	prompt := buildPrompt(toolName, req.ToolInput.RawString(), userMsgs)

	t0 := time.Now()
	dec, err := callLLM(&cfg.LLM, systemTemplate, prompt)
	elapsed := time.Since(t0)

	if err != nil {
		log.Printf("auto-guard: LLM error: %v (%.1fs)", err, elapsed.Seconds())
		return askDecision("LLM review unavailable, asking user")
	}

	log.Printf("auto-guard: LLM review %s in %v → %s", toolName, elapsed, dec.Decision)

	if dec.Decision == DecisionAllow {
		result := allowDecision("AI: " + dec.Reasoning)
		result.SystemMessage = "auto-guard: " + dec.Reasoning
		return result
	}

	return askDecision("AI: " + dec.Reasoning)
}

// ---- tier helpers ----

func isReadOnlyBash(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	for _, prefix := range readOnlyBashPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	base := strings.Fields(trimmed)
	if len(base) > 0 && readOnlyBashExact[base[0]] {
		return true
	}
	return false
}

func isTier2Bash(cmd string) bool {
	trimmed := strings.TrimSpace(cmd)
	for _, prefix := range tier2BashPrefixes {
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

func extractBashTarget(cmd string) string {
	// Simple extraction: last argument that looks like a path
	fields := strings.Fields(cmd)
	for i := len(fields) - 1; i >= 0; i-- {
		f := fields[i]
		if f == "" || strings.HasPrefix(f, "-") {
			continue
		}
		if strings.Contains(f, string(filepath.Separator)) || strings.HasSuffix(f, ".tex") || strings.HasSuffix(f, ".txt") {
			return f
		}
	}
	return ""
}

// ---- decision builders ----

func allowDecision(reason string) *Decision {
	return &Decision{
		Decision: DecisionAllow,
		Reason:   reason,
		HookSpecificOutput: &HookOutput{
			HookEventName:      "PreToolUse",
			PermissionDecision: "allow",
		},
	}
}

func askDecision(reason string) *Decision {
	return &Decision{
		Decision: DecisionAsk,
		Reason:   reason,
		HookSpecificOutput: &HookOutput{
			HookEventName:             "PreToolUse",
			PermissionDecision:        "ask",
			PermissionDecisionReason:  reason,
		},
	}
}
