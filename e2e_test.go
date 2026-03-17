package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2E(t *testing.T) {
	files, err := os.ReadDir("tests")
	if err != nil {
		t.Fatalf("Failed to read tests directory: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".lisp") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join("tests", file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", path, err)
			}

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

			// Prepare input (everything except the last line)
			input := strings.Join(lines[:len(lines)-1], "\n")

			cmd := exec.Command("go", "run", ".")
			cmd.Stdin = strings.NewReader(input + "\n")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			err = cmd.Run()
			if err != nil {
				t.Fatalf("Failed to run interpreter for %s: %v\nOutput: %s", path, err, out.String())
			}

			actualLines := strings.Split(strings.TrimSpace(out.String()), "\n")
			// Filter out empty lines if any
			var filteredActual []string
			for _, l := range actualLines {
				if strings.TrimSpace(l) != "" {
					filteredActual = append(filteredActual, strings.TrimSpace(l))
				}
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
