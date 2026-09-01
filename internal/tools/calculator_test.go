package tools

import (
	"testing"
)

func TestCalculator(t *testing.T) {
	calc := Calculator{}

	testCases := []struct {
		name        string
		operation   string
		input       []float64
		expected    float64
		expectError bool
	}{
		{"Addition", "add", []float64{1, 2, 3}, 6, false},
		{"Multiplication", "multiply", []float64{2, 3, 4}, 24, false},
		{"Subtraction", "subtract", []float64{10, 5, 2}, 3, false},
		{"Division", "divide", []float64{20, 4, 2}, 2.5, false},
		{"Modulus", "modulus", []float64{10, 3}, 1, false},
		{"Division by zero", "divide", []float64{10, 0}, 0, true},
		{"Modulus by zero", "modulus", []float64{10, 0}, 0, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			operation := testCase.operation
			numbers := testCase.input

			args := map[string]any{
				"operation": operation,
				"numbers":   numbers,
			}

			result, err := calc.Execute(args)

			if testCase.expectError {
				if err == nil {
					t.Errorf("Expected an error for %s, but got none", testCase.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", testCase.name, err)
				return
			}

			actual, ok := result.(float64)
			if !ok {
				t.Errorf("Expected result to be of type float64 for %s, but got %T", testCase.name, result)
			}

			if actual != testCase.expected {
				t.Errorf("For %s, expected %v but got %v", testCase.name, testCase.expected, actual)
			}
		})
	}
}
