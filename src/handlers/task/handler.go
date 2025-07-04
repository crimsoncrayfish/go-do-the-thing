package task

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
	"strconv"
	"time"

	templ_todo "go-do-the-thing/src/handlers/task/templ"

	fm "go-do-the-thing/src/models/forms"
	projects_service "go-do-the-thing/src/services/project"
	task_service "go-do-the-thing/src/services/task"
	templ_shared "go-do-the-thing/src/shared/templ"

	"github.com/a-h/templ"
)

type Handler struct {
	logger          slog.Logger
	task_service    task_service.TaskService
	project_service projects_service.ProjectService
}

var (
	source      = "TasksHandler"
	defaultForm = fm.NewDefaultTaskForm()
)

func SetupTodoHandler(
	taskService task_service.TaskService,
	projectService projects_service.ProjectService,
	router *http.ServeMux,
	mw_stack middleware.Middleware,
) {
	logger := slog.NewLogger(source)

	todoHandler := &Handler{
		task_service:    taskService,
		project_service: projectService,
		logger:          logger,
	}

	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.getItem)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.listItems)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItemStatus)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.createItem)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItem)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.deleteItem)))
	router.Handle("POST /todo/item/restore/{id}", mw_stack(http.HandlerFunc(todoHandler.restoreItem)))
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	form := fm.NewTaskForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Task Name")
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description", "Description")
	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		form.Errors["Project"] = err.Error()
	}
	project_id, err := strconv.ParseInt(project, 10, 64)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	due_date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     &due_date,
		ProjectId:   project_id,
	}
	if len(form.Errors) > 0 {
		if err := templ_todo.TaskFormContent(form, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to update task form")
		}
		return
	}

	new_id, err := h.task_service.CreateTask(current_user_id, project_id, name, description, &due_date)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to create task")
		return
	}

	// NOTE: Now handle everything
	taskView, err := h.task_service.GetTaskView(new_id, current_user_id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve newly created task")
		return
	}

	var form_template templ.Component
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		projects, err = h.project_service.GetAllProjectsForUser(current_user_id)
		if err != nil {
			defaultForm.Errors["Project"] = err.Error()
		}
		defaultForm.SetProject(taskView.ProjectId)
		form_template = templ_todo.TaskFormContent(defaultForm, models.ProjectListToMap(projects))
	} else {
		defaultForm.SetProject(taskView.ProjectId)
		form_template = templ_todo.TaskFormContent(defaultForm, map[int64]string{taskView.ProjectId: taskView.ProjectName})
	}

	err = templ_shared.RenderTempls(
		templ_todo.TaskItemCardOOB(taskView),
		templ_shared.NoDataRowOOB(true),
		form_template,
	).Render(r.Context(), w)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display new task")
		return
	}
}

type NoItemRowData struct {
	HideNoData bool
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid task ID provided")
		return
	}
	form := fm.NewTaskForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Task Name")
	if err != nil {
		form.Errors["name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description", "Description")

	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		form.Errors["Project"] = err.Error()
	}
	project_id, err := strconv.ParseInt(project, 10, 64)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["due_on"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["due_on"] = err.Error()
	}
	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     &date,
	}
	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		task, err := h.task_service.GetTaskView(id, current_user_id)
		if err != nil {
			errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve task for editing")
			return
		}
		if err := templ_todo.TaskItemContentWithErrors(task, models.ProjectListToMap(projects), form.Errors, true).Render(r.Context(), w); err != nil {
			errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display task edit form")
			return
		}
		return
	}

	// NOTE: Take action
	err = h.task_service.UpdateTask(
		current_user_id,
		id,
		project_id,
		name,
		description,
		&date,
		current_user_id,
	)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to update task")
		return
	}
	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve updated task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_todo.TaskItemContentOOBTargetList(task, models.ProjectListToMap(projects)),
		templ_todo.TaskCardFrontOOB(task),
	).Render(r.Context(), w)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display updated task")
		return
	}
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid task ID provided")
		return
	}
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// Check if it's a not found error
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code() == errors.ErrNotFound {
			errors.FrontendErrorNotFound(w, r, h.logger, err, "Task not found")
			return
		}
		// For other errors (permission, database, etc.), use internal server error
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve task")
		return
	}

	// NOTE: Take action
	formData := formDataFromTask(task)
	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		formData.Errors["Project"] = err.Error()
	}

	var template templ.Component
	source := r.URL.Query().Get("source")
	if source == "list" {
		template = templ_todo.TaskItemContentOOBTargetList(task, models.ProjectListToMap(projects))
	} else {
		contentType := r.Header.Get("accept")
		if contentType == "text/html" {
			template = templ_todo.TaskItem(task, models.ScreenTodo, formData, models.ProjectListToMap(projects))
		} else {
			template = templ_todo.TaskItemWithBody(task, models.ScreenTodo, formData, models.ProjectListToMap(projects))
		}
	}
	if err := template.Render(r.Context(), w); err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display task")
		return
	}
}

func (h *Handler) updateItemStatus(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid task ID provided")
		return
	}

	// NOTE: Act
	err = h.task_service.UpdateTaskStatus(current_user_id, id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to update task status")
		return
	}

	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: handle err types
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve updated task")
		return
	}
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
		if err != nil {
			defaultForm.Errors["Project"] = err.Error()
		}

		if err = templ_todo.TaskItemContent(task, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display updated task")
			return
		}
	} else {
		if err = templ_todo.TaskCardFront(task).Render(r.Context(), w); err != nil {
			errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display updated task")
			return
		}
	}
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Take action
	tasks, err := h.task_service.GetTaskViewList(current_user_id)
	if err != nil {
		// TODO: some user feedback here?
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve tasks")
		return
	}

	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		defaultForm.Errors["Project"] = err.Error()
	}

	// NOTE: Success zone
	var template templ.Component
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		template = templ_todo.TaskListPage(models.ScreenTodo, defaultForm, tasks, models.ProjectListToMap(projects))
	} else {
		template = templ_todo.TaskListWithBody(models.ScreenTodo, defaultForm, tasks, models.ProjectListToMap(projects))
	}
	if err = template.Render(r.Context(), w); err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display task list")
		return
	}
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid task ID provided")
		return
	}

	// NOTE: Take action
	err = h.task_service.DeleteTask(current_user_id, id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to delete task")
		return
	}

	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve deleted task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_shared.ToastMessage("Task Deleted", "warning"),
		templ_todo.TaskCardFront(task),
	).Render(r.Context(), w)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display task deletion result")
		return
	}
}

func (h *Handler) restoreItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "Invalid task ID provided")
		return
	}

	// NOTE: Take action
	err = h.task_service.RestoreTask(current_user_id, id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to restore task")
		return
	}

	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to retrieve restored task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_shared.ToastMessage("Task Restored", "success"),
		templ_todo.TaskCardFront(task),
	).Render(r.Context(), w)
	if err != nil {
		errors.FrontendErrorInternalServerError(w, r, h.logger, err, "Failed to display task restoration result")
		return
	}
}

func formDataFromTask(task *models.TaskView) fm.TaskForm {
	formData := fm.NewTaskForm()
	formData.Task.Name = task.Name
	formData.Task.Description = task.Description
	formData.Task.AssignedTo = task.AssignedTo
	formData.Task.DueDate = task.DueDate
	formData.Task.ProjectId = task.ProjectId
	formData.Task.ProjectName = task.ProjectName

	return formData
}
