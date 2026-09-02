package tools

import "testing"

func TestToolRegistry_Remove(t *testing.T) {
	testCases := []struct {
		name        string
		tools       []string
		removeTool  string
		expectError bool
	}{
		{"Remove existing tool", []string{"calculator", "shell"}, "calculator", false},
		{"Remove non-existing tool", []string{"calculator", "shell"}, "nonexistant", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tr := NewToolRegistry()

			for _, toolName := range testCase.tools {
				tr.Register(toolName, &Calculator{})
			}

			err := tr.Remove(testCase.removeTool)

			if testCase.expectError {
				if err == nil {
					t.Errorf("Expected error when removing tool '%s', but got none", testCase.removeTool)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error when removing tool '%s': %v", testCase.removeTool, err)
			}

			if tr.Has(testCase.removeTool) {
				t.Errorf("Expected tool '%s' to be removed", testCase.removeTool)
			}
		})
	}
}
