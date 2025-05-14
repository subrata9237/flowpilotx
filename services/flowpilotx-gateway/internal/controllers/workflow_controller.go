package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/flowpilotx/libs/logger"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/models"
	"github.com/flowpilotx/services/flowpilotx-gateway/internal/service"
	"github.com/gorilla/mux"
)

type WorkflowController struct {
	workflowService service.WorkflowService
	logger          logger.LoggerInterface
}

func NewWorkflowController(workflowService service.WorkflowService, logger logger.LoggerInterface) *WorkflowController {
	return &WorkflowController{
		workflowService: workflowService,
		logger:          logger,
	}
}

// CreateWorkflow godoc
// @Summary      Create a new workflow
// @Description  Create a new workflow with the provided schema
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        workflow  body      models.WorkflowSchema  true  "Workflow schema"
// @Success      201       {object}  models.WorkflowSchema
// @Failure      400       {string}  string  "Invalid request"
// @Failure      500       {string}  string  "Internal server error"
// @Router       /v1/workflows [post]
func (c *WorkflowController) CreateWorkflow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var workflow models.WorkflowSchema
		if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Invalid workflow request", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.workflowService.CreateWorkflow(r.Context(), &workflow); err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Failed to create workflow", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		models.WriteJSON(w, http.StatusCreated, workflow)
	}
}

// GetWorkflow godoc
// @Summary      Get workflow by ID
// @Description  Get workflow details by its ID
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Workflow ID"
// @Success      200  {object}  models.WorkflowSchema
// @Failure      500  {string}  string  "Internal server error"
// @Router       /v1/workflows/{id} [get]
func (c *WorkflowController) GetWorkflow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		workflow, err := c.workflowService.GetWorkflow(r.Context(), id)
		if err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Failed to get workflow", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		models.WriteJSON(w, http.StatusOK, workflow)
	}
}

// TriggerWorkflow godoc
// @Summary      Trigger a workflow
// @Description  Trigger a workflow execution by its ID
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        id    path      string                 true  "Workflow ID"
// @Param        input body      map[string]interface{} true  "Workflow input"
// @Success      200   {object}  map[string]string
// @Failure      400   {string}  string  "Invalid request"
// @Failure      500   {string}  string  "Internal server error"
// @Router       /v1/workflows/{id}/trigger [post]
func (c *WorkflowController) TriggerWorkflow() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Invalid workflow input", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		requestID, err := c.workflowService.TriggerWorkflow(r.Context(), id, input)
		if err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Failed to trigger workflow", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		models.WriteJSON(w, http.StatusOK, map[string]string{"request_id": requestID})
	}
}

// GetWorkflowRequest godoc
// @Summary      Get workflow request
// @Description  Get workflow request details by request ID
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Request ID"
// @Success      200  {object}  models.WorkflowRequest
// @Failure      500  {string}  string  "Internal server error"
// @Router       /v1/workflows/request/{id} [get]
func (c *WorkflowController) GetWorkflowRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		request, err := c.workflowService.GetWorkflowRequest(r.Context(), id)
		if err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Failed to get workflow request", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		models.WriteJSON(w, http.StatusOK, request)
	}
}

// GetActivityResults godoc
// @Summary      Get workflow activity results
// @Description  Get activity results for a workflow request
// @Tags         Workflows
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Request ID"
// @Success      200  {array}   models.ActivityResult
// @Failure      500  {string}  string  "Internal server error"
// @Router       /v1/workflows/request/{id}/activities [get]
func (c *WorkflowController) GetActivityResults() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		results, err := c.workflowService.GetActivityResults(r.Context(), id)
		if err != nil {
			c.logger.ErrorWithCtx(r.Context(), "Failed to get activity results", map[string]interface{}{"error": err.Error()})
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		models.WriteJSON(w, http.StatusOK, results)
	}
}
