package jev

import (
	"testing"
)

func TestShellPolicyEvaluate(t *testing.T) {
	var policy ShellPolicy

	testCases := []struct {
		name     string
		action   Action
		expected Decision
	}{
		{
			name: "shell tool",
			action: Action{
				Tool: "shell",
			},
			expected: Confirm,
		},
		{
			name: "non shell tool",
			action: Action{
				Tool: "calculator",
			},
			expected: Allow,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var result DecisionResult
			result = policy.Evaluate(testCase.action)

			if result.Decision != testCase.expected {
				t.Errorf("expected %v, got %v", testCase.expected, result.Decision)
			}
		})
	}
}
