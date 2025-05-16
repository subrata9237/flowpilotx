package activities

import (
	"fmt"
)

func (a *Activities) ExecuteAdd(input map[string]interface{}) (map[string]interface{}, error) {
	a.Logger.Info("Executing add activity", map[string]interface{}{
		"input": input,
	})

	x, ok := input["a"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: a must be a number")
		a.Logger.Error("Invalid input for add activity", map[string]interface{}{
			"error": err.Error(),
			"input": input,
		})
		return map[string]interface{}{}, err
	}
	y, ok := input["b"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: b must be a number")
		a.Logger.Error("Invalid input for add activity", map[string]interface{}{
			"error": err.Error(),
			"input": input,
		})
		return map[string]interface{}{}, err
	}

	result := x + y
	a.Logger.Info("Add activity completed", map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	})
	return map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	}, nil
}

func (a *Activities) ExecuteMultiply(input map[string]interface{}) (map[string]interface{}, error) {
	a.Logger.Info("Executing multiply activity", map[string]interface{}{
		"input": input,
	})

	x, ok := input["a"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: a must be a number")
		a.Logger.Error("Invalid input for multiply activity", map[string]interface{}{
			"error": err.Error(),
			"input": input,
		})
		return map[string]interface{}{}, err
	}
	y, ok := input["b"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: b must be a number")
		a.Logger.Error("Invalid input for multiply activity", map[string]interface{}{
			"error": err.Error(),
			"input": input,
		})
		return map[string]interface{}{}, err
	}

	result := x * y
	a.Logger.Info("Multiply activity completed", map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	})
	return map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	}, nil
}
