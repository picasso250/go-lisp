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
	expected []string // 每行输出的预期结果
}

func TestE2E(t *testing.T) {
	cases := []testCase{
		{
			name:  "Basic Arithmetic",
			input: "(+ 1 2) (* 3 4) (- 10 5) (/ 10 2)",
			expected: []string{"3", "12", "5", "5"},
		},
		{
			name:  "Factorial Recursion",
			input: "(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1)))))) (fact 5) (fact 6)",
			expected: []string{"120", "720"},
		},
		{
			name:  "List Operations",
			input: "(define x '(1 2 3)) (car x) (cdr x) (length (cons 0 x))",
			expected: []string{"1", "[2 3]", "4"},
		},
		{
			name:  "Let Scoping",
			input: "(define x 10) (let ((x 20)) x) x",
			expected: []string{"20", "10"},
		},
		{
			name:  "Begin Block",
			input: "(begin (define a 1) (define b 2) (+ a b))",
			expected: []string{"3"},
		},
		{
			name:  "Nested Conditionals",
			input: "(if (> 10 5) 'yes 'no) (if (< 10 5) 'yes 'no)",
			expected: []string{"yes", "no"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 执行 go run main.go
			cmd := exec.Command("go", "run", "main.go")
			cmd.Stdin = strings.NewReader(tc.input + "\n")
			
			var out bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &out

			err := cmd.Run()
			if err != nil {
				t.Fatalf("Failed to run interpreter: %v\nOutput: %s", err, out.String())
			}

			// 清理并分割输出
			actualLines := strings.Split(strings.TrimSpace(out.String()), "\n")
			
			// 验证输出行数
			if len(actualLines) != len(tc.expected) {
				t.Errorf("Expected %d lines of output, got %d. Output: %q", len(tc.expected), len(actualLines), actualLines)
				return
			}

			// 逐行匹配
			for i, exp := range tc.expected {
				actual := strings.TrimSpace(actualLines[i])
				if actual != exp {
					t.Errorf("Line %d: expected %q, got %q", i+1, exp, actual)
				}
			}
		})
	}
}
