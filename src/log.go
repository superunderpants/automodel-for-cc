package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	logMu      sync.Mutex
	logFile    *os.File
	logInited  bool
	classifyMu sync.Mutex // prevents interleaved output from concurrent hook calls
)

func initLog() {
	if logInited {
		return
	}
	logInited = true

	dir := configDir()
	os.MkdirAll(dir, 0755)
	path := filepath.Join(dir, "guard.log")

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	logFile = f

	multi := io.MultiWriter(f, os.Stderr)
	log.SetOutput(multi)
	log.SetFlags(log.Ltime)
}

func logf(format string, args ...interface{}) {
	initLog()
	logMu.Lock()
	defer logMu.Unlock()
	log.Printf(format, args...)
}

func logSection(title string) {
	line := fmt.Sprintf("──── %s ────", title)
	log.Println(line)
}

func logKV(k, v string) {
	if len(v) > 500 {
		v = v[:500] + "...(truncated)"
	}
	log.Printf("  %-10s %s", k+":", v)
}

func logDivider() {
	log.Println("─────────────────────────────")
}

func elapsedLog(start time.Time) string {
	return time.Since(start).Round(time.Millisecond).String()
}

// logBlock holds the lock for the entire block of a classify() call.
// All log* functions called within the returned func() are serialized.
func logBlock() func() {
	initLog()
	classifyMu.Lock()
	logMu.Lock()
	return func() {
		logMu.Unlock()
		classifyMu.Unlock()
	}
}

// ---- path helpers (used by guard.go) ----

func hasPathSep(s string) bool {
	return strings.Contains(s, "/") || strings.Contains(s, "\\") || strings.Contains(s, string(filepath.Separator))
}
