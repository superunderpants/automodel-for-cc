package main

import (
	"strings"
)

const systemTemplate = `You are a security classifier for a coding assistant.
Your job is to decide whether a pending tool operation is safe to auto-allow or should ask the user.
Respond with exactly one JSON object.`

func buildPrompt(toolName, input, userMessages string) string {
	var b strings.Builder
	b.WriteString("A coding assistant is about to execute a tool operation.\n")
	b.WriteString("Decide whether to auto-allow or ask the user.\n\n")

	// project context
	b.WriteString("## Project Context\n")
	b.WriteString("Root: " + detectProjectRoot() + "\n")
	b.WriteString("Name: " + projectName() + "\n")
	if claudeMD != "" {
		b.WriteString("Rules:\n---\n")
		b.WriteString(claudeMD)
		b.WriteString("\n---\n")
	}
	b.WriteString("\n")

	// location hint
	if input != "" {
		if isInsideProject(input) {
			b.WriteString("Location: INSIDE project\n\n")
		} else {
			b.WriteString("Location: OUTSIDE project\n\n")
		}
	}

	// user messages (reasoning-blind — only user messages, no assistant)
	if userMessages != "" {
		b.WriteString("## User Request (recent)\n")
		b.WriteString("---\n")
		b.WriteString(userMessages)
		b.WriteString("\n---\n\n")
	}

	// the operation
	b.WriteString("## Pending Action\n")
	b.WriteString("Tool: " + toolName + "\n")
	if len(input) > 1000 {
		input = input[:1000] + "..."
	}
	b.WriteString("Input: " + input + "\n\n")

	// decision
	b.WriteString("## Decision\n")
	b.WriteString(`Respond with exactly one JSON object:
{"decision": "<allow|uncertain>", "reasoning": "<prompt-safe summary>", "reasoning_long": "<3-4 sentence evidence summary>"}

- "allow" — the user clearly intended this action. Auto-approve silently.
- "uncertain" — not enough context to confirm user intent. Ask the user.
- The user request is your primary signal. If the user asked for this or it follows naturally, choose allow.
- Only choose uncertain when the action goes beyond what the user described.
- When in doubt, choose uncertain. The user will simply be prompted.`)

	return b.String()
}
