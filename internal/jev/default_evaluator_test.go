package jev

import (
	"testing"
)

func TestDefaultEvaluatorEvaluate(t *testing.T) {
	var evaluator *DefaultEvaluator
	evaluator = NewDefaultEvaluator()

	testCases := []struct {
		name     string
		action   Action
		expected Decision
	}{
		{
			name: "normal tool",
			action: Action{
				Tool: "calculator",
			},
			expected: Allow,
		},
		{
			name: "shell tool",
			action: Action{
				Tool: "shell",
			},
			expected: Confirm,
		},
		{
			name: "empty tool",
			action: Action{
				Tool: "",
			},
			expected: Deny,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var result DecisionResult
			result = evaluator.Evaluate(testCase.action)

			if result.Decision != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, result.Decision)
			}
		})
	}
}
