package activities

import (
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	"github.com/flowpilotx/libs/worker/constants"
)

type ActivityFunc func(map[string]interface{}) (map[string]interface{}, error)

type Activities struct {
	ActivityMap map[string]ActivityFunc
	Logger      logger.LoggerInterface
	MongoClient *mongodb.Client
}

func (a *Activities) RegisterActivities() {
	a.ActivityMap = make(map[string]ActivityFunc)

	// Add more detailed logging
	a.Logger.Info("Registering activities", nil)

	a.ActivityMap[constants.ActivityAdd] = a.ExecuteAdd
	a.ActivityMap[constants.ActivityMultiply] = a.ExecuteMultiply

	a.Logger.Info("Activities registered", map[string]interface{}{
		"activities": a.GetAvailableActivities(), // Use GetAvailableActivities instead of raw map
		"count":      len(a.ActivityMap),
	})
}

func (a *Activities) GetAvailableActivities() []string {
	activities := make([]string, 0, len(a.ActivityMap))
	for activity := range a.ActivityMap {
		activities = append(activities, activity)
	}
	return activities
}
