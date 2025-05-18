# flowpilotx

mongosh "mongodb://<subratasuperadmin>:<password123>@mongo-wf-storage-0.mongo-wf-storage.default.default.svc.cluster.local:27017"

//Test workflow to create workflow

 {
  "_id": {
    "$oid": "68290e1ebc86c8d258b0c835"
  },
  "queue": "flowpilotx-default-queue",
  "name": "math-workflow",
  "category": "calculation",
  "team": "data-team",
  "priority": "high",
  "type": "sequential",
  "description": "Basic math operations workflow",
  "activities": [
    {
      "id": "a1b2c3d4-e5f6-4321-8901-abcdef123456",
      "name": "Add",
      "type": "Add",
      "queue": "flowpilotx-add-queue",
      "input_schema": {
        "b": "number",
        "a": "number",
        "custom": {
          "factor": [
            100
          ]
        }
      },
      "output_schema": {
        "result": "number"
      },
      "config": {
        "async": false,
        "timeout": {
          "start_to_close": "30m0s",
          "schedule_to_start": "5m0s",
          "schedule_to_close": "1h0m0s",
          "heartbeat": "30s"
        }
      },
      "retry": {
        "initial_interval": "1s",
        "backoff_coefficient": 2,
        "max_interval": "1m0s",
        "max_attempts": 3,
        "non_retryable_error_types": [
          "InvalidInputError",
          "NumberFormatError"
        ]
      },
      "attempt": 0,
      "status": ""
    },
    {
      "id": "b2c3d4e5-f6a1-5432-9012-bcdef1234567",
      "name": "Multiply",
      "type": "Multiply",
      "queue": "flowpilotx-multiply-queue",
      "input_schema": {
        "a": "$.a1b2c3d4-e5f6-4321-8901-abcdef123456.result",
        "b": {
          "$ref": "$.a1b2c3d4-e5f6-4321-8901-abcdef123456.custom.factor[0]"
        }
      },
      "output_schema": {
        "result": "number"
      },
      "config": {
        "async": false,
        "timeout": {
          "start_to_close": "30m0s",
          "schedule_to_start": "5m0s",
          "schedule_to_close": "1h0m0s",
          "heartbeat": "30s"
        }
      },
      "retry": {
        "initial_interval": "1s",
        "backoff_coefficient": 2,
        "max_interval": "1m0s",
        "max_attempts": 3,
        "non_retryable_error_types": [
          "InvalidInputError",
          "NumberFormatError",
          "DivideByZeroError"
        ]
      },
      "attempt": 0,
      "status": ""
    }
  ],
  "dag": {
    "a1b2c3d4-e5f6-4321-8901-abcdef123456": [],
    "b2c3d4e5-f6a1-5432-9012-bcdef1234567": [
      "a1b2c3d4-e5f6-4321-8901-abcdef123456"
    ]
  },
  "created_at": {
    "$date": "2025-05-17T22:30:54.636Z"
  },
  "updated_at": {
    "$date": "2025-05-17T22:30:54.636Z"
  },
  "status": "ACTIVE",
  "version": 1
}

//To test workflow
id:68290e1ebc86c8d258b0c835
{
    "a": 5,
    "b": 3,
    "custom": {"factor":[100]}
  }