package activity

import (
	"fmt"

	workerModel "github.com/flowpilotx/libs/worker/model"
)

func (a *Activity) Add(activity *workerModel.ActivityDefinition) (*workerModel.ActivityDefinition, error) {
	a.Logger.Info("Executing add activity", map[string]interface{}{
		"input": activity,
	})

	x, ok := activity.InputSchema["a"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: a must be a number")
		a.Logger.Error("Invalid input for add activity", map[string]interface{}{
			"error": err.Error(),
			"input": activity,
		})
		return activity, err
	}
	y, ok := activity.InputSchema["b"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: b must be a number")
		a.Logger.Error("Invalid input for add activity", map[string]interface{}{
			"error": err.Error(),
			"input": activity,
		})
		return activity, err
	}

	result := x + y
	activity.OutputSchema = map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	}

	a.Logger.Info("Add activity completed", map[string]interface{}{
		"activity_id": activity.ID,
		"inputs": map[string]interface{}{
			"a": x,
			"b": y,
		},
		"result": result,
	})

	return activity, nil
}

func (a *Activity) Multiply(activity *workerModel.ActivityDefinition) (*workerModel.ActivityDefinition, error) {
	a.Logger.Info("Executing multiply activity", map[string]interface{}{
		"input": activity,
	})

	x, ok := activity.InputSchema["a"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: a must be a number")
		a.Logger.Error("Invalid input for multiply activity", map[string]interface{}{
			"error": err.Error(),
			"input": activity,
		})
		return activity, err
	}
	y, ok := activity.InputSchema["b"].(float64)
	if !ok {
		err := fmt.Errorf("invalid input: b must be a number")
		a.Logger.Error("Invalid input for multiply activity", map[string]interface{}{
			"error": err.Error(),
			"input": activity,
		})
		return activity, err
	}

	result := x * y
	a.Logger.Info("Multiply activity completed", map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	})
	activity.OutputSchema = map[string]interface{}{
		"a":      x,
		"b":      y,
		"result": result,
	}
	return activity, nil
}
