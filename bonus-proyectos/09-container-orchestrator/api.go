package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ContainerStatus is the JSON representation of a container's status for the API.
type ContainerStatus struct {
	Name        string        `json:"name"`
	State       ContainerState `json:"state"`
	PID         int           `json:"pid"`
	ExitCode    int           `json:"exit_code"`
	Healthy     bool          `json:"healthy"`
	StartCount  int           `json:"start_count"`
	LastStarted string        `json:"last_started,omitempty"`
	LastStopped string        `json:"last_stopped,omitempty"`
	RestartPolicy string      `json:"restart_policy"`
}

// APIResponse is the standard JSON envelope for all API responses.
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// APIServer provides a minimal HTTP REST API for monitoring and controlling
// the orchestrator and its managed containers.
type APIServer struct {
	orchestrator *Orchestrator
	logger       *Logger
	port         int
}

// NewAPIServer creates an API server bound to the given orchestrator and port.
func NewAPIServer(orchestrator *Orchestrator, logger *Logger, port int) *APIServer {
	return &APIServer{
		orchestrator: orchestrator,
		logger:       logger,
		port:         port,
	}
}

// Start begins listening for HTTP requests. It should be called as a goroutine.
func (api *APIServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.handleHealth)
	mux.HandleFunc("/containers", api.handleContainers)
	mux.HandleFunc("/containers/start", api.handleStartContainer)
	mux.HandleFunc("/containers/stop", api.handleStopContainer)
	mux.HandleFunc("/logs", api.handleLogs)

	addr := fmt.Sprintf(":%d", api.port)
	api.logger.Log("api", "info", fmt.Sprintf("REST API listening on %s", addr))

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}

// handleHealth responds to GET /health with the health status of all containers.
func (api *APIServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status:  "error",
			Message: "method not allowed",
		})
		return
	}

	containers := api.orchestrator.GetAllContainers()
	statuses := api.orchestrator.health.GetLatestStatus(containers)

	// Check if all are healthy.
	allHealthy := true
	for _, s := range statuses {
		if !s.Healthy {
			allHealthy = false
			break
		}
	}

	httpStatus := http.StatusOK
	msg := "all containers healthy"
	if !allHealthy {
		httpStatus = http.StatusServiceUnavailable
		msg = "one or more containers unhealthy"
	}

	writeJSON(w, httpStatus, APIResponse{
		Status:  msg,
		Data:    statuses,
	})
}

// handleContainers responds to GET /containers with detailed status of all containers.
func (api *APIServer) handleContainers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status:  "error",
			Message: "method not allowed",
		})
		return
	}

	containers := api.orchestrator.GetAllContainers()
	var statuses []ContainerStatus

	for _, c := range containers {
		c.mu.RLock()
		cs := ContainerStatus{
			Name:          c.Config.Name,
			State:         c.State,
			PID:           c.PID,
			ExitCode:      c.ExitCode,
			Healthy:       c.Healthy,
			StartCount:    c.StartCount,
			RestartPolicy: c.Config.RestartPolicy,
		}
		if !c.LastStarted.IsZero() {
			cs.LastStarted = c.LastStarted.Format(time.RFC3339)
		}
		if !c.LastStopped.IsZero() {
			cs.LastStopped = c.LastStopped.Format(time.RFC3339)
		}
		c.mu.RUnlock()
		statuses = append(statuses, cs)
	}

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "ok",
		Data:   statuses,
	})
}

// handleStartContainer responds to POST /containers/start?name=<name>
// to start a stopped/failed container.
func (api *APIServer) handleStartContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status:  "error",
			Message: "method not allowed",
		})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Status:  "error",
			Message: "missing 'name' query parameter",
		})
		return
	}

	if err := api.orchestrator.StartContainer(name); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	api.logger.Log("api", "info", fmt.Sprintf("container %s started via API", name))
	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "ok",
		Message: fmt.Sprintf("container %s started", name),
	})
}

// handleStopContainer responds to POST /containers/stop?name=<name>
// to gracefully stop a running container.
func (api *APIServer) handleStopContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status:  "error",
			Message: "method not allowed",
		})
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		writeJSON(w, http.StatusBadRequest, APIResponse{
			Status:  "error",
			Message: "missing 'name' query parameter",
		})
		return
	}

	if err := api.orchestrator.StopContainer(name); err != nil {
		writeJSON(w, http.StatusInternalServerError, APIResponse{
			Status:  "error",
			Message: err.Error(),
		})
		return
	}

	api.logger.Log("api", "info", fmt.Sprintf("container %s stopped via API", name))
	writeJSON(w, http.StatusOK, APIResponse{
		Status:  "ok",
		Message: fmt.Sprintf("container %s stopped", name),
	})
}

// handleLogs responds to GET /logs with all aggregated log entries.
func (api *APIServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Status:  "error",
			Message: "method not allowed",
		})
		return
	}

	entries := api.logger.GetEntries()

	writeJSON(w, http.StatusOK, APIResponse{
		Status: "ok",
		Data:   entries,
	})
}

// writeJSON serializes the response as JSON and writes it to the HTTP response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}