package tools

import "testing"

func TestToolRegistry(t *testing.T) {

	testCases := []struct {
		name        string
		toolName    string
		tool        Tool
		expectError bool
	}{
		{"Register and retrieve calculator", "calculator", &Calculator{}, false},
		{"Calculator not registered", "nonexistant", nil, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			tr := NewToolRegistry()

			if testCase.tool != nil {
				tr.Register(testCase.toolName, testCase.tool)
			}

			actualTool, err := tr.Get(testCase.toolName)

			if testCase.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", testCase.name)
				}

				if actualTool != nil {
					t.Errorf("Expected no tool for %s, but got one", testCase.name)
				}

				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", testCase.name, err)
			}

			if actualTool == nil {
				t.Errorf("Expected tool for %s, but got nil", testCase.name)
			}
		})
	}
}
