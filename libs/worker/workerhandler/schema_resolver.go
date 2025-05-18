// libs/worker/workerhandler/schema_resolver.go
package workerhandler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/worker/model"
)

type SchemaResolver struct {
	logger     logger.LoggerInterface
	activities map[string]*model.ActivityDefinition
}

func NewSchemaResolver(logger logger.LoggerInterface) *SchemaResolver {
	return &SchemaResolver{
		logger:     logger,
		activities: make(map[string]*model.ActivityDefinition),
	}
}

func (r *SchemaResolver) ResolveInputSchema(activity *model.ActivityDefinition) (map[string]interface{}, error) {
	resolvedInputs := make(map[string]interface{})

	r.logger.Info("Starting input schema resolution", map[string]interface{}{
		"activity_id":  activity.ID,
		"input_schema": activity.InputSchema,
	})

	for key, value := range activity.InputSchema {
		resolved, err := r.resolveValue(value)
		if err != nil {
			r.logger.Error("Failed to resolve input", map[string]interface{}{
				"activity_id": activity.ID,
				"key":         key,
				"value":       value,
				"error":       err.Error(),
			})
			return nil, fmt.Errorf("failed to resolve input '%s': %v", key, err)
		}
		resolvedInputs[key] = resolved
	}

	r.logger.Info("Completed input schema resolution", map[string]interface{}{
		"activity_id":     activity.ID,
		"resolved_inputs": resolvedInputs,
	})
	return resolvedInputs, nil
}

func (r *SchemaResolver) resolveValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		if strings.HasPrefix(v, "$.") {
			r.logger.Info("Resolving string reference", map[string]interface{}{
				"reference": v,
			})
			return r.resolvePathReference(v)
		}
		return v, nil

	case map[string]interface{}:
		if ref, ok := v["$ref"].(string); ok {
			r.logger.Info("Resolving object reference", map[string]interface{}{
				"reference": ref,
			})
			return r.resolvePathReference(ref)
		}
		return v, nil

	default:
		return v, nil
	}
}

func (r *SchemaResolver) resolvePathReference(path string) (interface{}, error) {
	path = strings.TrimPrefix(path, "$.")
	parts := strings.SplitN(path, ".", 2)

	if len(parts) != 2 {
		r.logger.Error("Invalid reference path", map[string]interface{}{
			"path": path,
		})
		return nil, fmt.Errorf("invalid reference path: %s", path)
	}

	activityID, nestedPath := parts[0], parts[1]

	sourceActivity, exists := r.activities[activityID]
	if !exists {
		r.logger.Error("Referenced activity not found", map[string]interface{}{
			"activity_id": activityID,
			"path":        path,
		})
		return nil, fmt.Errorf("referenced activity '%s' not found", activityID)
	}

	r.logger.Debug("Starting nested path resolution", map[string]interface{}{
		"activity_id": activityID,
		"nested_path": nestedPath,
		"schema":      sourceActivity.OutputSchema,
	})
	value, err := r.resolveNestedValue(sourceActivity.OutputSchema, nestedPath, activityID)
	if err != nil {
		return r.resolveNestedValue(sourceActivity.InputSchema, nestedPath, activityID)
	}
	return value, nil
}

func (r *SchemaResolver) resolveNestedValue(data map[string]interface{}, path string, activityID string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		r.logger.Debug("Resolving path segment", map[string]interface{}{
			"activity_id": activityID,
			"segment":     part,
			"depth":       i,
			"remaining":   strings.Join(parts[i:], "."),
		})

		// Handle array indexing
		if arrayIndex := r.parseArrayIndex(part); arrayIndex >= 0 {
			arrayName := strings.Split(part, "[")[0]
			value, err := r.resolveArrayValue(current, arrayName, arrayIndex, activityID)
			if err != nil {
				return nil, err
			}

			// If this is the last part, return the value
			if i == len(parts)-1 {
				return value, nil
			}

			// Otherwise, continue traversing if it's a map
			if nextMap, ok := value.(map[string]interface{}); ok {
				current = nextMap
				continue
			}
			return nil, fmt.Errorf("cannot traverse through non-object array element at: %s", part)
		}

		// Handle object field access
		value, exists := current[part]
		if !exists {
			r.logger.Error("Field not found in path", map[string]interface{}{
				"activity_id": activityID,
				"field":       part,
				"full_path":   path,
			})
			return nil, fmt.Errorf("field '%s' not found in path: %s", part, path)
		}

		// If this is the last part, return the value
		if i == len(parts)-1 {
			r.logger.Info("Successfully resolved nested reference", map[string]interface{}{
				"activity_id": activityID,
				"path":        path,
				"value":       value,
			})
			return value, nil
		}

		// Continue traversing if it's a map
		if nextMap, ok := value.(map[string]interface{}); ok {
			current = nextMap
		} else {
			r.logger.Error("Cannot traverse non-object value", map[string]interface{}{
				"activity_id": activityID,
				"field":       part,
				"type":        fmt.Sprintf("%T", value),
			})
			return nil, fmt.Errorf("cannot traverse through non-object value at: %s (type: %T)", part, value)
		}
	}

	return nil, fmt.Errorf("invalid path resolution")
}

func (r *SchemaResolver) parseArrayIndex(part string) int {
	if idx := strings.Index(part, "["); idx >= 0 && strings.HasSuffix(part, "]") {
		numStr := part[idx+1 : len(part)-1]
		if num, err := strconv.Atoi(numStr); err == nil && num >= 0 {
			return num
		}
	}
	return -1
}

func (r *SchemaResolver) resolveArrayValue(data map[string]interface{}, arrayName string, index int, activityID string) (interface{}, error) {
	array, ok := data[arrayName].([]interface{})
	if !ok {
		r.logger.Error("Not an array", map[string]interface{}{
			"activity_id": activityID,
			"field":       arrayName,
		})
		return nil, fmt.Errorf("not an array: %s", arrayName)
	}

	if index >= len(array) {
		r.logger.Error("Array index out of bounds", map[string]interface{}{
			"activity_id": activityID,
			"array":       arrayName,
			"index":       index,
			"length":      len(array),
		})
		return nil, fmt.Errorf("array index out of bounds: %s[%d]", arrayName, index)
	}

	return array[index], nil
}
