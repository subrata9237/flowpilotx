package activity

import (
	"time"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	workerModel "github.com/flowpilotx/libs/worker/model"
	"go.temporal.io/sdk/client"
)

type Activity struct {
	Logger         logger.LoggerInterface
	MongoClient    *mongodb.Client
	TemporalClient client.Client
}

func (a *Activity) prepareFinalActivity(activity *workerModel.ActivityDefinition, err error) (*workerModel.ActivityDefinition, error) {
	now := time.Now()
	activity.EndTime = &now
	duration := now.Sub(*activity.StartTime)
	activity.Duration = duration.String()
	if err != nil {
		activity.Status = "failed"
		activity.Error = err.Error()
	} else {
		activity.Status = "completed"
	}
	return activity, nil
}
