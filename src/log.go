package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logMu    sync.Mutex
	logFile  *os.File
	logInited bool
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

	// Also tee to stderr so console has visibility
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
	initLog()
	logMu.Lock()
	defer logMu.Unlock()
	line := fmt.Sprintf("──── %s ────", title)
	log.Println(line)
}

func logKV(k, v string) {
	if len(v) > 500 {
		v = v[:500] + "...(truncated)"
	}
	initLog()
	logMu.Lock()
	defer logMu.Unlock()
	log.Printf("  %-10s %s", k+":", v)
}

func logDivider() {
	initLog()
	logMu.Lock()
	defer logMu.Unlock()
	log.Println("─────────────────────────────")
}

func elapsedLog(start time.Time) string {
	return time.Since(start).Round(time.Millisecond).String()
}
