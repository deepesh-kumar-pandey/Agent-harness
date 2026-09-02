package tools

import "testing"

func TestToolRegistry_Has(t *testing.T) {

	testCases := []struct {
		name      string
		toolName  string
		tool      Tool
		expectHas bool
	}{
		{"Tool is registered", "calculator", &Calculator{}, true},
		{"Tool is not registered", "nonexistant", nil, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			tr := NewToolRegistry()

			if testCase.tool != nil {
				tr.Register(testCase.toolName, testCase.tool)
			}

			actualHas := tr.Has(testCase.toolName)

			if actualHas != testCase.expectHas {
				t.Errorf("Expected Has(%q) to be %v, but got %v", testCase.toolName, testCase.expectHas, actualHas)
			}
		})
	}
}
