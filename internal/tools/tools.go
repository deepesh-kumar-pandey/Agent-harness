package tools

import "fmt"

type Tool interface {
	Name() string
	Description() string
	Execute(args map[string]any) (any, error)
}

func (c Calculator) Name() string {
	return "calculator"
}

func (c Calculator) Description() string {
	return "Performs basic arithmetic operations"
}

func (c Calculator) Execute(args map[string]any) (any, error) {
	operation, ok := args["operation"].(string)

	if !ok {
		return nil, fmt.Errorf("operation must be a string")
	}

	numbers, ok := args["numbers"].([]float64)

	if !ok {
		return nil, fmt.Errorf("numbers must have decimal values")
	}

	switch operation {
	case "add":
		return c.Add(numbers...)
	case "multiply":
		return c.Multiply(numbers...)
	case "subtract":
		return c.Subtract(numbers...)
	case "divide":
		return c.Divide(numbers...)
	case "modulus":
		if len(numbers) != 2 {
			return nil, fmt.Errorf("modulus operation requires exactly 2 numbers")
		}
		return c.Modulus(numbers[0], numbers[1])
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}
