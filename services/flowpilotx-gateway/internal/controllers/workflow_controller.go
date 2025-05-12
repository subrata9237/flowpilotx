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
