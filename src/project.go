package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ---- project root detection ----

var (
	projectRoot     string
	projectRootOnce sync.Once
	claudeMD        string
)

var projectMarkers = []string{
	"CLAUDE.md",
	".git",
	"package.json",
	"go.mod",
	"Cargo.toml",
	"setup.py",
	"pyproject.toml",
	"Makefile",
	"CMakeLists.txt",
}

func detectProjectRoot() string {
	projectRootOnce.Do(func() {
		projectRoot = detectOnce()
		// read CLAUDE.md
		path := filepath.Join(projectRoot, "CLAUDE.md")
		if data, err := os.ReadFile(path); err == nil {
			// trim to a reasonable size
			s := string(data)
			if len(s) > 2000 {
				s = s[:2000] + "..."
			}
			claudeMD = s
		}
	})
	return projectRoot
}

func detectOnce() string {
	// 1. git
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		root := strings.TrimSpace(string(out))
		if root != "" {
			return filepath.ToSlash(root)
		}
	}
	// 2. walk up looking for project markers
	cwd, _ := os.Getwd()
	cur, _ := filepath.Abs(cwd)
	cur = filepath.ToSlash(cur)
	for {
		for _, m := range projectMarkers {
			if _, err := os.Stat(filepath.Join(cur, m)); err == nil {
				return cur
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	// 3. fallback to CWD
	return filepath.ToSlash(cwd)
}

func projectName() string {
	return filepath.Base(detectProjectRoot())
}

func isInsideProject(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	abs = filepath.ToSlash(abs)
	root := filepath.ToSlash(detectProjectRoot())
	return strings.HasPrefix(abs, root)
}
