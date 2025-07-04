package task

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
		}
		return
	}

	new_id, err := h.task_service.CreateTask(current_user_id, project_id, name, description, &due_date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err, "failed to create task with error")
		return
	}

	// NOTE: Now handle everything
	taskView, err := h.task_service.GetTaskView(new_id, current_user_id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err, "failed to get newly created task with id %d", new_id)
		return
	}
	err = templ_shared.RenderTempls(
		templ_todo.TaskItemCardOOB(taskView),
		templ_shared.NoDataRowOOB(true),
	).Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err, "failed to render templates for new task")
		return
	}

	source := r.URL.Query().Get("source")
	if source == "task_page" {
		projects, err = h.project_service.GetAllProjectsForUser(current_user_id)
		if err != nil {
			defaultForm.Errors["Project"] = err.Error()
		}
		defaultForm.SetProject(taskView.ProjectId)
		if err := templ_todo.TaskFormContent(defaultForm, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.logger.Error(err, "failed to render task form")
			return
		}
	} else {
		defaultForm.SetProject(taskView.ProjectId)
		if err := templ_todo.TaskFormContent(defaultForm, map[int64]string{taskView.ProjectId: taskView.ProjectName}).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			h.logger.Error(err, "failed to render task form")
			return
		}
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.logger.Error(err, "not authenticated")
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
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if err := templ_todo.TaskItemContentWithErrors(task, models.ProjectListToMap(projects), form.Errors, true).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task form content")
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		h.logger.Error(err, "failed to update task with id %d", id)
		if err := templ_shared.ToastMessage("Failed to update task", "error").Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to render toast")
			return
		}
	}
	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to get updated task with id %d", task.Id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templ_todo.TaskItemContentOOBTargetList(task, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render new task row with id %d", task.Id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templ_todo.TaskCardFrontOOB(task).Render(r.Context(), w)
	if err != nil {
		h.logger.Error(err, "failed to render task list item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: Handle not found?
		h.logger.Error(err, "failed to get task with id %d", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: Take action
	formData := formDataFromTask(task)
	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		formData.Errors["Project"] = err.Error()
	}

	source := r.URL.Query().Get("source")
	if source == "list" {
		if err = templ_todo.TaskItemContentOOBTargetList(task, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task item with id %d", id)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		contentType := r.Header.Get("accept")
		if contentType == "text/html" {
			if err = templ_todo.TaskItem(task, models.ScreenTodo, formData, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
				h.logger.Error(err, "failed to render task item with id %d", id)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if err = templ_todo.TaskItemWithBody(task, models.ScreenTodo, formData, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
				h.logger.Error(err, "failed to render task item with id %d", id)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
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
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Act
	err = h.task_service.UpdateTaskStatus(current_user_id, id)
	if err != nil {
		h.logger.Error(err, "failed to toggle task status for task with id %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Success zone
	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: handle err types
		h.logger.Error(err, "failed to get task with id %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
		if err != nil {
			defaultForm.Errors["Project"] = err.Error()
		}

		if err = templ_todo.TaskItemContent(task, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task list item")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskCardFront(task).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task list item")
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		defaultForm.Errors["Project"] = err.Error()
	}

	// NOTE: Success zone
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_todo.TaskListPage(models.ScreenTodo, defaultForm, tasks, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskListWithBody(models.ScreenTodo, defaultForm, tasks, models.ProjectListToMap(projects)).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Take action
	err = h.task_service.DeleteTask(current_user_id, id)
	if err != nil {
		assert.NoError(err, source, "failed to delete todo item")
		return
	}

	// NOTE: Success zone
	if err := templ_shared.ToastMessage("Task Deleted", "warning").Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render toast")
		return
	}

	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: handle err types
		h.logger.Error(err, "failed to get task with id %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		h.logger.Debug("TODO:WHAT")
	} else {
		if err = templ_todo.TaskCardFront(task).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task list item")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Take action
	err = h.task_service.RestoreTask(current_user_id, id)
	if err != nil {
		assert.NoError(err, source, "failed to restore todo item %d", id)
		return
	}

	// NOTE: Success zone
	if err := templ_shared.ToastMessage("Task Restored", "success").Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render toast")
		// TODO: what should happen if the fetch fails after create
		return
	}

	task, err := h.task_service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: handle err types
		h.logger.Error(err, "failed to get task with id %d", id)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		h.logger.Debug("TODO:WHAT")
	} else {
		if err = templ_todo.TaskCardFront(task).Render(r.Context(), w); err != nil {
			h.logger.Error(err, "failed to render task list item")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
