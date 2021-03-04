package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/MarioCarrion/todo-api/internal"
)

const uuidRegEx string = `[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}`

//go:generate counterfeiter -o resttesting/task_service.gen.go . TaskService

// TaskService ...
type TaskService interface {
	Create(ctx context.Context, description string, priority internal.Priority, dates internal.Dates) (internal.Task, error)
	Task(ctx context.Context, id string) (internal.Task, error)
	Update(ctx context.Context, id string, description string, priority internal.Priority, dates internal.Dates, isDone bool) error
}

// TaskHandler ...
type TaskHandler struct {
	svc TaskService
}

// NewTaskHandler ...
func NewTaskHandler(svc TaskService) *TaskHandler {
	return &TaskHandler{
		svc: svc,
	}
}

// Register connects the handlers to the router.
func (t *TaskHandler) Register(r *mux.Router) {
	r.HandleFunc("/tasks", t.create).Methods(http.MethodPost)
	r.HandleFunc(fmt.Sprintf("/tasks/{id:%s}", uuidRegEx), t.task).Methods(http.MethodGet)
	r.HandleFunc(fmt.Sprintf("/tasks/{id:%s}", uuidRegEx), t.update).Methods(http.MethodPut)
}

// Task is an activity that needs to be completed within a period of time.
type Task struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

// CreateTasksRequest defines the request used for creating tasks.
type CreateTasksRequest struct {
	Description string   `json:"description"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

// CreateTasksResponse defines the response returned back after creating tasks.
type CreateTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(w, "invalid request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	task, err := t.svc.Create(r.Context(), req.Description, req.Priority.Convert(), req.Dates.Convert())
	if err != nil {
		renderErrorResponse(w, "create failed", http.StatusInternalServerError)
		return
	}

	renderResponse(w,
		&CreateTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
			},
		},
		http.StatusCreated)
}

// ReadTasksResponse defines the response returned back after searching one task.
type ReadTasksResponse struct {
	Task Task `json:"task"`
}

func (t *TaskHandler) task(w http.ResponseWriter, r *http.Request) {
	id, _ := mux.Vars(r)["id"] // NOTE: Safe to ignore error, because it's always defined.

	task, err := t.svc.Task(r.Context(), id)
	if err != nil {
		// XXX: Differentiating between NotFound and Internal errors will be covered in future episodes.
		renderErrorResponse(w, "find failed", http.StatusInternalServerError)
		return
	}

	renderResponse(w,
		&ReadTasksResponse{
			Task: Task{
				ID:          task.ID,
				Description: task.Description,
				Priority:    NewPriority(task.Priority),
				Dates:       NewDates(task.Dates),
			},
		},
		http.StatusOK)
}

// UpdateTasksRequest defines the request used for updating a task.
type UpdateTasksRequest struct {
	Description string   `json:"description"`
	IsDone      bool     `json:"is_done"`
	Priority    Priority `json:"priority"`
	Dates       Dates    `json:"dates"`
}

func (t *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTasksRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		renderErrorResponse(w, "invalid request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	id, _ := mux.Vars(r)["id"] // NOTE: Safe to ignore error, because it's always defined.

	err := t.svc.Update(r.Context(), id, req.Description, req.Priority.Convert(), req.Dates.Convert(), req.IsDone)
	if err != nil {
		// XXX: Differentiating between NotFound and Internal errors will be covered in future episodes.
		renderErrorResponse(w, "update failed", http.StatusInternalServerError)
		return
	}

	renderResponse(w, &struct{}{}, http.StatusOK)
}
