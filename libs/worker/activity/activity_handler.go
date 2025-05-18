package activity

import (
	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/libs/mongodb"
	"go.temporal.io/sdk/client"
)

type Activity struct {
	Logger         logger.LoggerInterface
	MongoClient    *mongodb.Client
	TemporalClient client.Client
}
