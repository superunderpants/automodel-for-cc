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
	homeDir, _ := os.UserHomeDir()
	homeDir = filepath.ToSlash(homeDir)

	// 1. git
	if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
		root := strings.TrimSpace(string(out))
		if root != "" {
			root = filepath.ToSlash(root)
			if root != homeDir {
				return root
			}
		}
	}
	// 2. walk up looking for project markers (skip home dir)
	cwd, _ := os.Getwd()
	cur, _ := filepath.Abs(cwd)
	cur = filepath.ToSlash(cur)
	for {
		if cur != homeDir {
			for _, m := range projectMarkers {
				if _, err := os.Stat(filepath.Join(cur, m)); err == nil {
					return cur
				}
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

func normalizePath(p string) string {
	// Convert Cygwin/MSYS2/Git Bash paths like /c/foo to C:/foo on Windows
	if len(p) >= 3 && p[0] == '/' && p[2] == '/' && p[1] >= 'a' && p[1] <= 'z' {
		p = strings.ToUpper(p[1:2]) + ":" + p[2:]
	}
	return p
}

func isInsideProject(path string) bool {
	path = normalizePath(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	abs = filepath.ToSlash(abs)
	root := filepath.ToSlash(detectProjectRoot())
	return strings.HasPrefix(abs, root)
}
