// libs/worker/workerhandler/schema_resolver.go
package workerhandler

import (
	"fmt"
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

	for key, value := range activity.InputSchema {
		resolved, err := r.resolveValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve input '%s': %v", key, err)
		}
		resolvedInputs[key] = resolved
	}

	return resolvedInputs, nil
}

func (r *SchemaResolver) resolveValue(value interface{}) (interface{}, error) {
	switch v := value.(type) {
	case string:
		// Handle direct reference (e.g., "$.0.result")
		if strings.HasPrefix(v, "$.") {
			return r.resolvePathReference(v)
		}
		return v, nil

	case map[string]interface{}:
		// Handle object reference (e.g., {"$ref": "$.0.b"})
		if ref, ok := v["$ref"].(string); ok {
			return r.resolvePathReference(ref)
		}
		return v, nil

	default:
		return v, nil
	}
}

func (r *SchemaResolver) resolvePathReference(path string) (interface{}, error) {
	// Remove the leading "$."
	path = strings.TrimPrefix(path, "$.")

	// Split into activity ID and field path
	parts := strings.SplitN(path, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference path: %s", path)
	}

	activityID, fieldPath := parts[0], parts[1]

	// Get source activity
	sourceActivity, exists := r.activities[activityID]
	if !exists {
		return nil, fmt.Errorf("referenced activity '%s' not found", activityID)
	}

	// Get value from output schema
	value, exists := sourceActivity.OutputSchema[fieldPath]
	if !exists {
		return nil, fmt.Errorf("field '%s' not found in activity '%s' output", fieldPath, activityID)
	}

	return value, nil
}
