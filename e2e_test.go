package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	testsDir, err := filepath.Abs("tests")
	if err != nil {
		t.Fatalf("Failed to get absolute path for tests: %v", err)
	}
	stdlibPath := filepath.Join(wd, "stdlib.lisp")

	files, err := os.ReadDir(testsDir)
	if err != nil {
		t.Fatalf("Failed to read tests directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".lisp") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(testsDir, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", path, err)
			}

			// Create temporary directory for each test
			tmpDir := t.TempDir()
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("Failed to change to temporary directory: %v", err)
			}
			defer os.Chdir(wd)

			lines := strings.Split(strings.TrimSpace(string(content)), "\n")
			if len(lines) == 0 {
				t.Fatalf("Test file %s is empty", path)
			}

			lastLine := lines[len(lines)-1]
			if !strings.HasPrefix(lastLine, "; EXPECT") {
				t.Fatalf("Test file %s missing ; EXPECT comment on last line", path)
			}

			expectedJSON := strings.TrimPrefix(lastLine, "; EXPECT")
			var expected []string
			if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
				t.Fatalf("Failed to parse expected output in %s: %v", path, err)
			}

			// 现在设置 defer，因为它能访问已经解析出的 expected
			defer func() {
				if r := recover(); r != nil {
					if len(expected) > 0 && expected[0] == "PANIC" {
						// 预期的 panic
						return
					}
					t.Fatalf("Test %s panicked unexpectedly: %v", path, r)
				}
			}()

			input := strings.Join(lines[:len(lines)-1], "\n")

			env := standardEnv()
			if stdlib, err := os.ReadFile(stdlibPath); err == nil {
				evalString(io.Discard, string(stdlib), env, true)
			}

			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			var buffer bytes.Buffer
			evalString(&buffer, input, env, false)

			w.Close()
			os.Stdout = old

			var stdoutBuffer bytes.Buffer
			io.Copy(&stdoutBuffer, r)

			combinedOutput := buffer.String() + stdoutBuffer.String()

			actualLines := strings.Split(strings.TrimSpace(combinedOutput), "\n")
			var filteredActual []string
			for _, l := range actualLines {
				if strings.TrimSpace(l) != "" {
					filteredActual = append(filteredActual, strings.TrimSpace(l))
				}
			}

			// 如果标记为 PANIC 且没有 panic 发生，也是一种失败
			if len(expected) > 0 && expected[0] == "PANIC" {
				t.Errorf("%s: Expected panic but it didn't happen", path)
				return
			}

			if len(filteredActual) != len(expected) {
				t.Errorf("%s: Expected %d lines, got %d. Output: %q", path, len(expected), len(filteredActual), filteredActual)
				return
			}
			for i, exp := range expected {
				if filteredActual[i] != exp {
					t.Errorf("%s: Line %d: expected %q, got %q", path, i+1, exp, filteredActual[i])
				}
			}
		})
	}
}
