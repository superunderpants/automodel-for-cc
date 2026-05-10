package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ---- output types (Claude Code PreToolUse hook protocol) ----

const (
	DecisionAllow = "allow"
	DecisionAsk   = "ask"
)

type Decision struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason,omitempty"`
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
	if len(text) > 4000 {
		text = text[len(text)-4000:]
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, `"type":"tool_result"`) {
			continue
		}
		if strings.Contains(trimmed, `"role":"user"`) ||
			strings.Contains(trimmed, `"role": "user"`) ||
			strings.HasPrefix(trimmed, "user:") ||
			strings.HasPrefix(trimmed, "User:") {
			lines = append(lines, trimmed)
		}
	}
	if len(lines) == 0 {
		if len(text) > 2000 {
			text = text[len(text)-2000:]
		}
		return text
	}
	return strings.Join(lines, "\n")
}

// ---- main classification logic ----

func classify(req *HookRequest) *Decision {
	unlock := logBlock()
	defer unlock()

	toolName := req.CanonicalTool()
	cmd := req.ToolInput.Command()
	filePath := req.ToolInput.FilePath()

	logDivider()
	logSection("INCOMING")
	logKV("tool", toolName)
	if cmd != "" {
		logKV("command", cmd)
	}
	if filePath != "" {
		logKV("file", filePath)
	}
	if req.TranscriptPath != "" {
		logKV("transcript", req.TranscriptPath)
	}

	// Tier 1: always-allow tools
	if tier1Tools[toolName] {
		logSection("RESULT - TIER 1")
		logKV("decision", "allow")
		logKV("reason", "safe read-only tool")
		return allowDecision("tier 1: safe read-only tool")
	}

	// Tier 1: Bash with read-only patterns
	if toolName == "Bash" {
		if isReadOnlyBash(cmd) {
			logSection("RESULT - TIER 1")
			logKV("decision", "allow")
			logKV("reason", "read-only bash command")
			return allowDecision("tier 1: read-only bash command")
		}
		// Tier 2: Bash safe writes inside project
		if isTier2Bash(cmd) {
			target := extractBashTarget(cmd)
			if target != "" && isInsideProject(target) {
				logSection("RESULT - TIER 2")
				logKV("decision", "allow")
				logKV("reason", "safe bash write inside project")
				return allowDecision("tier 2: safe bash write inside project")
			}
		}
	}

	// Tier 2: Write/Edit inside project
	if tier2Tools[toolName] {
		if filePath != "" && isInsideProject(filePath) {
			logSection("RESULT - TIER 2")
			logKV("decision", "allow")
			logKV("reason", "safe edit inside project")
			return allowDecision("tier 2: safe edit inside project")
		}
	}

	// Tier 3: LLM review
	logSection("TIER 3 - LLM REVIEW")
	logKV("status", "loading config...")

	cfg, err := LoadConfig()
	if err != nil {
		logKV("error", err.Error())
		return askDecision("config error, asking user")
	}

	logKV("url", cfg.LLM.BaseURL)
	logKV("model", cfg.LLM.Model)

	userMsgs := readUserMessages(req.TranscriptPath)
	prompt := buildPrompt(toolName, req.ToolInput.RawString(), userMsgs)

	logKV("context", userMsgs)
	logKV("question", prompt)

	t0 := time.Now()
	logKV("status", "calling LLM...")
	dec, err := callLLM(&cfg.LLM, systemTemplate, prompt)

	if err != nil {
		logKV("error", fmt.Sprintf("%v (elapsed: %s)", err, elapsedLog(t0)))
		return askDecision("LLM review unavailable, asking user")
	}

	logSection("LLM RESPONSE")
	logKV("elapsed", elapsedLog(t0))
	logKV("decision", dec.Decision)
	logKV("reasoning", dec.Reasoning)

	if dec.Decision == DecisionAllow {
		return allowDecision(dec.Reasoning)
	}
	return askDecision(dec.Reasoning)
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
	fields := strings.Fields(cmd)
	for i := len(fields) - 1; i >= 0; i-- {
		f := fields[i]
		if f == "" || strings.HasPrefix(f, "-") {
			continue
		}
		if hasPathSep(f) || strings.HasSuffix(f, ".tex") || strings.HasSuffix(f, ".txt") {
			return f
		}
	}
	return ""
}

// ---- decision builders ----

func allowDecision(reason string) *Decision {
	return &Decision{
		HookEventName:      "PreToolUse",
		PermissionDecision: "allow",
	}
}

func askDecision(reason string) *Decision {
	return &Decision{
		HookEventName:            "PreToolUse",
		PermissionDecision:       "ask",
		PermissionDecisionReason: reason,
	}
}
