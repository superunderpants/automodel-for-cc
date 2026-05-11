package main

import (
	"encoding/json"
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
	// Only look at the tail of the transcript for recent context
	if len(text) > 16000 {
		text = text[len(text)-16000:]
	}

	var userTexts []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		// Only user-turn messages, not assistant or system
		if t, _ := msg["type"].(string); t != "user" {
			continue
		}
		// Extract text content blocks (skip tool_result blocks)
		inner := msg["message"]
		if inner == nil {
			continue
		}
		innerMsg, ok := inner.(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := innerMsg["content"].([]interface{})
		for _, block := range content {
			b, ok := block.(map[string]interface{})
			if !ok {
				continue
			}
			if t, _ := b["type"].(string); t != "text" {
				continue
			}
			if txt, _ := b["text"].(string); txt != "" {
				userTexts = append(userTexts, txt)
			}
		}
	}
	if len(userTexts) == 0 {
		return ""
	}
	result := strings.Join(userTexts, "\n")
	if len(result) > 2000 {
		result = result[len(result)-2000:]
	}
	return result
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

	// Tier 1 + Tier 2: Bash — only if no shell operators (redirects, pipes, chains, etc.)
	if toolName == "Bash" && !hasShellOperators(cmd) {
		if isReadOnlyBash(cmd) {
			logSection("RESULT - TIER 1")
			logKV("decision", "allow")
			logKV("reason", "read-only bash command")
			return allowDecision("tier 1: read-only bash command")
		}
		// Tier 2: Bash with whitelisted write commands inside project
		// All non-flag file arguments must be inside the project root.
		if isTier2Bash(cmd) {
			targets := extractBashTargets(cmd)
			if len(targets) > 0 {
				allInside := true
				for _, t := range targets {
					if !isInsideProject(t) {
						allInside = false
						break
					}
				}
				if allInside {
					logSection("RESULT - TIER 2")
					logKV("decision", "allow")
					logKV("reason", "safe bash write inside project")
					return allowDecision("tier 2: safe bash write inside project")
				}
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

func hasShellOperators(cmd string) bool {
	for _, op := range []string{">", ">>", "<", "|", "&&", "||", ";", "&", "`", "$("} {
		if strings.Contains(cmd, op) {
			return true
		}
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

func extractBashTargets(cmd string) []string {
	fields := strings.Fields(cmd)
	if len(fields) < 2 {
		return nil
	}
	var targets []string
	for i := 1; i < len(fields); i++ {
		f := fields[i]
		if strings.HasPrefix(f, "-") {
			continue
		}
		targets = append(targets, stripQuotes(f))
	}
	return targets
}

func stripQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
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
