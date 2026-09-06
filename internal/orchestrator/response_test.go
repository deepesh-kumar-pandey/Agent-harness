package orchestrator

import (
	"encoding/json"
	"fmt"
	"testing"
)

// ─────────────────────────────────────────────
// Test AgentResponse
// ─────────────────────────────────────────────

func TestAgentResponse(t *testing.T) {

	fmt.Println("Starting AgentResponse tests...")

	testCases := []struct {
		name            string
		input           string
		expectError     bool
		expectTool      bool
		expectedContent string
		expectedTool    string
	}{
		{
			name: "Normal response",
			input: `{
				"content": "Hello"
			}`,
			expectError:     false,
			expectTool:      false,
			expectedContent: "Hello",
		},
		{
			name: "Response with tool call",
			input: `{
				"content": "",
				"tool_call": {
					"tool": "calculator",
					"args": {
						"operation": "add",
						"numbers": [10, 20]
					}
				}
			}`,
			expectError:  false,
			expectTool:   true,
			expectedTool: "calculator",
		},
		{
			name:        "Invalid JSON",
			input:       `{"content": "Hello"`,
			expectError: true,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {

			fmt.Printf("Running test: %s\n", testCase.name)

			var response AgentResponse

			err := json.Unmarshal(
				[]byte(testCase.input),
				&response,
			)

			// ─────────────────────────────────────
			// Error validation
			// ─────────────────────────────────────

			if testCase.expectError && err == nil {
				t.Fatalf("Expected an error, but got nil")
			}

			if !testCase.expectError && err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// ─────────────────────────────────────
			// Success validation
			// ─────────────────────────────────────

			if !testCase.expectError {

				if response.Content != testCase.expectedContent {
					t.Fatalf(
						"Expected content %q, got %q",
						testCase.expectedContent,
						response.Content,
					)
				}

				if testCase.expectTool {

					if response.ToolCall == nil {
						t.Fatalf("Expected tool call, but got nil")
					}

					if response.ToolCall.Tool != testCase.expectedTool {
						t.Fatalf(
							"Expected tool %q, got %q",
							testCase.expectedTool,
							response.ToolCall.Tool,
						)
					}

					fmt.Printf(
						"Tool correctly parsed: %s\n",
						response.ToolCall.Tool,
					)

				} else {

					if response.ToolCall != nil {
						t.Fatalf("Expected no tool call")
					}

					fmt.Printf(
						"Response correctly parsed: %s\n",
						response.Content,
					)
				}
			}

			// ─────────────────────────────────────
			// Expected error
			// ─────────────────────────────────────

			if testCase.expectError {
				fmt.Printf(
					"Error correctly returned: %v\n",
					err,
				)
			}
		})
	}

	fmt.Println("AgentResponse tests completed!")
}
