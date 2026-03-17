package main

import (
	"testing"
)

func TestEdgeCases(t *testing.T) {
	// Test parse with empty tokens (main.go L102-L104)
	res, rest := parse([]string{})
	if res != nil || rest != nil {
		t.Errorf("Expected nil, nil for empty tokens")
	}

	// Test eval with nil (main.go L151-L153 or L229)
	resEval := eval(nil, nil)
	if resEval != nil {
		t.Errorf("Expected nil for eval(nil)")
	}

	// Test tokenize with a trailing word (main.go L95-L97)
	tokens := tokenize("abc")
	if len(tokens) != 1 || tokens[0] != "abc" {
		t.Errorf("Expected [abc], got %v", tokens)
	}

	// Test eval with an empty List (main.go L151-L153)
	resEmptyList := eval(List{}, nil)
	if resEmptyList != nil {
		t.Errorf("Expected nil for eval(List{})")
	}
}
