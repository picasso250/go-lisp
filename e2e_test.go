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
		{
			name:  "Advanced List Ops",
			input: "(append '(1 2) '(3 4)) (nth '(10 20 30) 1) (reverse '(1 2 3))",
			expected: []string{"[1 2 3 4]", "20", "[3 2 1]"},
		},
		{
			name:  "Higher Order Map",
			input: "(map (lambda (x) (* x 2)) '(1 2 3))",
			expected: []string{"[2 4 6]"},
		},
		{
			name:  "Advanced String Ops",
			input: "(string-length \"hello\") (string-contains? \"hello\" \"ell\") (string-replace \"abcabc\" \"a\" \"x\")",
			expected: []string{"5", "true", "xbcxbc"},
		},
		{
			name:  "Functional Filter & Reduce",
			input: "(filter (lambda (x) (> x 1)) '(1 2 3)) (reduce (lambda (acc x) (+ acc x)) '(1 2 3) 0)",
			expected: []string{"[2 3]", "6"},
		},
		{
			name:  "File Operations",
			input: "(write-file \"tmp.txt\" \"hello file\") (read-file \"tmp.txt\")",
			expected: []string{"true", "hello file"},
		},
		{
			name:  "Stdlib Functions",
			input: "(not true) (not false) (even? 4) (even? 5) (foldl + '(1 2 3) 10)",
			expected: []string{"false", "true", "true", "false", "16"},
		},
		{
			name:  "Short-circuit Logic",
			input: "(and true true false) (and true true) (or false false true) (or false false)",
			expected: []string{"false", "true", "true", "false"},
		},
		{
			name:  "Python Math & Conversions",
			input: "(abs -50) (pow 2 10) (str 123) (int \"456\") (bool 0) (bool 1) (bool '())",
			expected: []string{"50", "1024", "123", "456", "false", "true", "false"},
		},
		{
			name:  "Range and Zip",
			input: "(range 3) (zip '(a b) '(1 2)) (sum (range 11))",
			expected: []string{"[0 1 2]", "[[a 1] [b 2]]", "55"},
		},
		{
			name:  "Enumerate",
			input: "(enumerate '(apple banana))",
			expected: []string{"[[0 apple] [1 banana]]"},
		},
		{
			name:  "Hex and Bin",
			input: "(bin 10) (hex 255)",
			expected: []string{"0b1010", "0xff"},
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
