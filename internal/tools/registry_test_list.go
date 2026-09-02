package tools

import "testing"

func TestToolRegistry_List(t *testing.T) {
	testCases := []struct {
		name          string
		tools         []string
		expectedcount int
	}{
		{"Empty registry", []string{}, 0},
		{"One tool", []string{"calculator"}, 1},
		{"Multiple tools", []string{"calculator", "shell", "search"}, 3},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tr := NewToolRegistry()

			for _, toolName := range testCase.tools {
				tr.Register(toolName, &Calculator{}) // Using Calculator as a placeholder tool
			}

			actualCount := len(tr.List())

			if actualCount != testCase.expectedcount {
				t.Errorf("Expected %d tools in registry, but got %d", testCase.expectedcount, actualCount)
			}
		})
	}
}
