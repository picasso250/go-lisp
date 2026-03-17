package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

type testCase struct {
	name     string
	input    string
	expected []string
}

func TestE2E(t *testing.T) {
	cases := []testCase{
		{
			name:  "Big Integers",
			input: "(* 100000000000000000000 100000000000000000000)",
			expected: []string{"10000000000000000000000000000000000000000"},
		},
		{
			name:  "Type Conversions & Predicates",
			input: "(integer? 1) (float? 1.0) (integer 1.5) (float 2) (float? (float 1))",
			expected: []string{"true", "true", "1", "2", "true"},
		},
		{
			name:  "String Escapes",
			input: "(concat \"line1\\nline2\" \"\\\"quote\\\"\")",
			expected: []string{"line1", "line2\"quote\""},
		},
		{
			name:  "String Functions",
			input: "(string-trim \"  hello  \") (string-split \"a,b,c\" \",\")",
			expected: []string{"hello", "[a b c]"},
		},
		{
			name:  "Strict Arithmetic (Integer Only)",
			input: "(+ 1 2) (- 10 5)",
			expected: []string{"3", "5"},
		},
		{
			name:  "Strict Arithmetic (Float Only)",
			input: "(+ 1.1 2.2)",
			expected: []string{"3.3000000000000003"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("go", "run", ".")
			cmd.Stdin = strings.NewReader(tc.input + "\n")
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out
			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to run interpreter: %v\nOutput: %s", err, out.String())
			}
			actualLines := strings.Split(strings.TrimSpace(out.String()), "\n")
			if len(actualLines) != len(tc.expected) {
				t.Errorf("Expected %d lines, got %d. Output: %q", len(tc.expected), len(actualLines), actualLines)
				return
			}
			for i, exp := range tc.expected {
				if strings.TrimSpace(actualLines[i]) != exp {
					t.Errorf("Line %d: expected %q, got %q", i+1, exp, actualLines[i])
				}
			}
		})
	}
}
