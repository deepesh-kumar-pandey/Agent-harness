package jev

import (
	"testing"
)

func TestBasicPolicyEvaluate(t *testing.T) {
	var policy BasicPolicy

	testCases := []struct {
		name     string
		action   Action
		expected Decision
	}{
		{
			name: "empty tool",
			action: Action{
				Tool: "",
			},
			expected: Deny,
		},
		{
			name: "valid tool",
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
