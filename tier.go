package main

// ---- Tier 1: always allow (no AI) ----

var tier1Tools = map[string]bool{
	"Read":             true,
	"Glob":             true,
	"Grep":             true,
	"Task":             true,
	"TaskCreate":       true,
	"TaskUpdate":       true,
	"TaskGet":          true,
	"TaskList":         true,
	"TaskOutput":       true,
	"TaskStop":         true,
	"AskUserQuestion":  true,
	"EnterPlanMode":    true,
	"ExitPlanMode":     true,
	"CronList":         true,
	"Skill":            true,
}

// ---- Tier 1: always allow for read-only Bash patterns ----

var readOnlyBashPrefixes = []string{
	"ls ", "dir ", "cat ", "head ", "tail ", "wc ", "du ", "df ",
	"which ", "where ", "type ", "echo ", "date ", "pwd", "whoami",
	"uname ", "hostname ", "printenv", "env", "id", "groups",
	"find ", "tree ", "stat ", "file ", "grep ", "awk ", "sed ",
	"sort ", "uniq ", "cut ", "tr ", "tee ", "diff ", "cmp ",
	"go version", "go env", "rustc --version", "cargo --version",
	"python --version", "python3 --version", "node --version",
	"npm --version", "pip --version", "git status", "git log",
	"git diff", "git branch", "git show", "git blame",
	"latexmk", "pdflatex", "xelatex", "lualatex", "bibtex",
	"pdftotext", "pdfinfo", "kpsewhich",
}

var readOnlyBashExact = map[string]bool{
	"clear":     true,
	"reset":     true,
	"history":   true,
	"pwd":       true,
	"dir":       true,
	"ls":        true,
	"whoami":    true,
	"hostname":  true,
	"date":      true,
}

// ---- Tier 2: safe tools when inside project ----

var tier2Tools = map[string]bool{
	"Write":         true,
	"Edit":          true,
	"MultiEdit":     true,
	"NotebookEdit":  true,
}

// ---- known safe Bash tools that write (project context validates them) ----

var tier2BashPrefixes = []string{
	"mkdir ", "touch ", "cp ", "mv ", "rename ",
	"ln ", "chmod ", "chown ",
}

// ---- Tier 3: everything else goes to LLM ----

var tier3Tools = map[string]bool{
	"Bash":       true,
	"WebFetch":   true,
	"WebSearch":  true,
	// mcp__* tools
}
