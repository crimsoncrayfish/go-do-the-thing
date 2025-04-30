package task

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
	"strconv"
	"time"

	templ_todo "go-do-the-thing/src/handlers/task/templ"

	fm "go-do-the-thing/src/models/forms"
	task_service "go-do-the-thing/src/services/task"
	templ_shared "go-do-the-thing/src/shared/templ"
)

type Handler struct {
	logger  slog.Logger
	service task_service.TaskService
}

var (
	activeScreens models.NavBarObject
	source        = "TasksHandler"
	defaultForm   = fm.NewDefaultTaskForm()
)

func SetupTodoHandler(
	taskService task_service.TaskService,
	router *http.ServeMux,
	mw_stack middleware.Middleware,
) {
	logger := slog.NewLogger(source)

	activeScreens = models.NavBarObject{ActiveScreens: models.ActiveScreens{IsTodoList: true}}
	todoHandler := &Handler{
		service: taskService,
		logger:  logger,
	}

	router.Handle("GET /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.getItem)))
	router.Handle("GET /todo/items", mw_stack(http.HandlerFunc(todoHandler.listItems)))
	router.Handle("POST /todo/item/status/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItemStatus)))
	router.Handle("POST /todo/item", mw_stack(http.HandlerFunc(todoHandler.createItem)))
	router.Handle("POST /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.updateItem)))
	router.Handle("DELETE /todo/item/{id}", mw_stack(http.HandlerFunc(todoHandler.deleteItem)))
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	// TODO: what to do here
	if err != nil {
		// TODO: Handle properly
		http.Error(w, err.Error(), http.StatusUnauthorized)
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

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(due_date),
		Project:     project_id,
	}
	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := templ_todo.TaskFormContent("Create", form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
		}
		return
	}

	new_id, err := h.service.CreateTask(current_user_id, project_id, name, description, database.NewSqliteTime(due_date))
	if err != nil {
		// TODO: what should happen if the fetch fails after create
		return
	}

	// NOTE: Now handle everything
	taskView, err := h.service.GetTaskView(new_id, current_user_id)
	if err != nil {
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_todo.TaskRowOOB(taskView).Render(r.Context(), w); err != nil {
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		// if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_todo.TaskFormContent("Create", defaultForm).Render(r.Context(), w); err != nil {
		// TODO: what should happen if rendering fails
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
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	//
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: some user feedback here?
		http.Error(w, err.Error(), http.StatusBadRequest)
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
	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(date),
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_todo.TaskFormContent("Update", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to render task form")
			return
		}
		return
	}

	// NOTE: Take action
	if err = h.service.UpdateTask(
		current_user_id,
		id,
		project_id,
		name,
		description,
		database.NewSqliteTime(date),
		current_user_id,
	); err != nil {
		h.logger.Error(err, "failed to update task")
		form.Errors["Task"] = "failed to update task"
		if err := templ_todo.TaskFormContent("Update", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	// NOTE: Success zone
	task, err := h.service.GetTaskView(id, current_user_id)
	if err != nil {
		assert.NoError(err, source, "failed to get updated task")
		// TODO: what should happen if the fetch fails after update
		return
	}

	if err := templ_todo.TaskItemContentOOB(task).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render new task row with id %d", task.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}

	formData := formDataFromItemNoValidation(task)
	if err := templ_todo.TaskFormContent("Update", formData).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render form content after update")
		// TODO: what should happen if the fetch fails after create
		return
	}
}

type ItemPageModel struct {
	Task   models.TaskView
	NavBar models.NavBarObject
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	// TODO: what to do here
	if err != nil {
		// TODO: Handle properly
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		// TODO: Handle error by type
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task, err := h.service.GetTaskView(id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to get todo tasks")
		// TODO: Handle error by type
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: Take action
	formData := formDataFromItemNoValidation(task)

	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_todo.TaskItem(task, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the item page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskItemWithBody(task, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the item page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) updateItemStatus(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	// TODO: what to do here
	if err != nil {
		// TODO: Handle properly
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Validate data
	err = h.service.UpdateTaskStatus(id, current_user_id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to toggle task status")
		// TODO: handle err types
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Success zone
	task, err := h.service.GetTaskView(id, current_user_id)
	if err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to get todo item")
		// TODO: handle err types
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = templ_todo.TaskRowContent(task).Render(r.Context(), w); err != nil {
		// TODO: what to do here
		h.logger.Error(err, "failed to render task list item")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type ListModel struct {
	Tasks    []models.TaskView
	NavBar   models.NavBarObject
	FormData fm.TaskForm
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	// TODO: what to do here
	if err != nil {
		// TODO: Handle properly
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// NOTE: Take action
	tasks, err := h.service.GetTaskViewList(current_user_id)
	if err != nil {
		// TODO: some user feedback here?
		h.logger.Error(err, "failed to get todo tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: Success zone
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_todo.TaskList(activeScreens, defaultForm, tasks).Render(r.Context(), w); err != nil {
			// TODO: Should this panic
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_todo.TaskListWithBody(activeScreens, defaultForm, tasks).Render(r.Context(), w); err != nil {
			// TODO: Should this panic
			h.logger.Error(err, "Failed to execute template for the item list page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	// TODO: what to do here
	if err != nil {
		// TODO: Handle properly
		http.Error(w, err.Error(), http.StatusUnauthorized)
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
	err = h.service.DeleteTask(current_user_id, id)
	if err != nil {
		assert.NoError(err, source, "failed to delete todo item")
		return
	}

	// NOTE: Success zone
	hasData, err := h.service.GetTaskCount(current_user_id)
	if err != nil {
		assert.NoError(err, source, "failed to update ui")
		return
	}

	if err := templ_shared.NoDataRowOOB(hasData > 0).Render(r.Context(), w); err != nil {
		// if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		assert.NoError(err, source, "failed to render no data row")
		// TODO: what should happen if the fetch fails after create
		return
	}
}

func formDataFromItemNoValidation(task *models.TaskView) fm.TaskForm {
	formData := fm.NewTaskForm()
	formData.Task.Name = task.Name
	formData.Task.Description = task.Description
	formData.Task.AssignedTo = task.AssignedTo
	formData.Task.DueDate = task.DueDate

	return formData
}
