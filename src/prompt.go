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

	// location hint — only for tools where input is a clean file path
	if input != "" && (toolName == "Write" || toolName == "Edit" || toolName == "MultiEdit" || toolName == "NotebookEdit" || toolName == "Read") {
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

- "allow" — the action is reasonably safe and consistent with development work. Auto-approve silently.
- "uncertain" — the action is clearly destructive, targets sensitive system paths, or contradicts stated intent. Ask the user.
- The assistant is a coding tool. Assume its operations serve a legitimate development purpose.
- Remote connections (ssh, scp), HTTP requests (curl, wget), package managers (npm, pip, apt), and build tools are normal development operations. Allow them unless the specific payload is obviously malicious (e.g., curl | sh, downloading and executing untrusted binaries).
- Do not block an action just because it *could* be dangerous. Block only when the evidence of harm is clear.
- The user request in the context below is your primary signal. If it matches the pending action, prefer allow.`)

	return b.String()
}
