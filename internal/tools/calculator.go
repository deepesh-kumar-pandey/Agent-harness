package tools

import (
	"fmt"
	"math"
)

type Calculator struct{}

func (c Calculator) Add(numbers ...float64) (float64, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("no numbers provided")
	}

	var sum float64

	for _, number := range numbers {
		sum += number
	}

	return sum, nil
}

func (c Calculator) Multiply(numbers ...float64) (float64, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("no numbers provided")
	}

	product := 1.0

	for _, number := range numbers {
		product *= number
	}

	return product, nil
}

func (c Calculator) Subtract(numbers ...float64) (float64, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("no numbers provided")
	}

	result := numbers[0]
	for _, number := range numbers[1:] {
		result -= number
	}
	return result, nil
}

func (c Calculator) Divide(numbers ...float64) (float64, error) {
	if len(numbers) == 0 {
		return 0, fmt.Errorf("no numbers provided")
	}

	result := numbers[0]
	for _, number := range numbers[1:] {
		if number == 0 {
			return 0, fmt.Errorf("division by zero")
		}

		result /= number
	}
	return result, nil
}

func (c Calculator) Modulus(a, b float64) (float64, error) {
	if b == 0 {
		return 0, fmt.Errorf("division by zero")
	}

	return math.Mod(a, b), nil
}

func (c Calculator) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type": "string",
			},
			"numbers": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "number",
				},
			},
		},
		"required": []string{
			"operation",
			"numbers",
		},
	}
}
