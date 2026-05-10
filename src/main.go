package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// ---- hook protocol types ----

type HookRequest struct {
	ToolName       string    `json:"tool_name"`
	ToolInput      ToolInput `json:"tool_input"`
	TranscriptPath string    `json:"transcript_path"`
}

type ToolInput struct {
	raw      map[string]interface{}
}

func (t ToolInput) Command() string {
	if v, ok := t.raw["command"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (t ToolInput) FilePath() string {
	for _, key := range []string{"file_path", "path", "notebook_path"} {
		if v, ok := t.raw[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func (t ToolInput) Content() string {
	if v, ok := t.raw["content"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (t ToolInput) RawString() string {
	if cmd := t.Command(); cmd != "" {
		return cmd
	}
	if fp := t.FilePath(); fp != "" {
		return fp
	}
	if c := t.Content(); c != "" {
		return c
	}
	// fallback: first non-empty string value
	for _, v := range t.raw {
		if s, ok := v.(string); ok && s != "" {
			if len(s) > 200 {
				return s[:200] + "..."
			}
			return s
		}
	}
	return ""
}

func (t *ToolInput) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &t.raw)
}

// ---- tool name normalization ----

func (r *HookRequest) CanonicalTool() string {
	name := strings.TrimSpace(r.ToolName)
	// mcp__ tools
	if strings.HasPrefix(name, "mcp__") {
		return "mcp__"
	}
	// known aliases
	switch name {
	case "MultiEdit", "multi_edit", "Multi_Edit":
		return "MultiEdit"
	case "NotebookEdit", "notebook_edit", "Notebook_Edit":
		return "NotebookEdit"
	}
	return name
}

// ---- main ----

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("auto-guard: ")

	// Read stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		writeDecision(askDecision("failed to read stdin"))
		return
	}

	var req HookRequest
	if err := json.Unmarshal(data, &req); err != nil {
		writeDecision(askDecision(fmt.Sprintf("failed to parse input: %v", err)))
		return
	}

	// Classify
	decision := classify(&req)
	writeDecision(decision)
}

func writeDecision(d *Decision) {
	wrapper := map[string]interface{}{
		"hookSpecificOutput": d,
	}
	out, err := json.Marshal(wrapper)
	if err != nil {
		fmt.Fprintf(os.Stderr, "auto-guard: marshal error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
