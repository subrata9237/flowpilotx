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

	activityID, fieldPath := parts[0], parts[1]

	sourceActivity, exists := r.activities[activityID]
	if !exists {
		r.logger.Error("Referenced activity not found", map[string]interface{}{
			"activity_id": activityID,
			"path":        path,
		})
		return nil, fmt.Errorf("referenced activity '%s' not found", activityID)
	}
	r.logger.Info("Resolved activity", map[string]interface{}{
		"activity_id":   activityID,
		"output_schema": sourceActivity.OutputSchema,
	})
	value, exists := sourceActivity.OutputSchema[fieldPath]
	if !exists {
		r.logger.Error("Field not found in activity output", map[string]interface{}{
			"activity_id": activityID,
			"field_path":  fieldPath,
			"path":        path,
		})
		return nil, fmt.Errorf("field '%s' not found in activity '%s' output", fieldPath, activityID)
	}

	r.logger.Info("Successfully resolved reference", map[string]interface{}{
		"activity_id": activityID,
		"field_path":  fieldPath,
		"value":       value,
	})
	return value, nil
}
