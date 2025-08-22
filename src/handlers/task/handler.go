// Package task provides the apis for the tasks
package task

import (
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"net/http"
	"slices"
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
	taskService    task_service.TaskService
	projectService projects_service.ProjectService
	logger         *logger.Logger
}

const source = "TasksHandler"

func SetupTodoHandler(
	taskService task_service.TaskService,
	projectService projects_service.ProjectService,
	router *http.ServeMux,
	mwStack middleware.Middleware,
) {
	log := logger.NewLogger(source)
	todoHandler := &Handler{
		taskService:    taskService,
		projectService: projectService,
		logger:         log,
	}

	router.Handle("GET /todo/item/{id}", mwStack(http.HandlerFunc(todoHandler.getItem)))
	router.Handle("GET /todo/items", mwStack(http.HandlerFunc(todoHandler.listItems)))
	router.Handle("GET /tasks/lazy", mwStack(http.HandlerFunc(todoHandler.listItemsLazy)))
	router.Handle("POST /tasks/search", mwStack(http.HandlerFunc(todoHandler.search)))
	router.Handle("POST /todo/item/status/{id}", mwStack(http.HandlerFunc(todoHandler.updateItemStatus)))
	router.Handle("POST /todo/item", mwStack(http.HandlerFunc(todoHandler.createItem)))
	router.Handle("POST /todo/item/{id}", mwStack(http.HandlerFunc(todoHandler.updateItem)))
	router.Handle("DELETE /todo/item/{id}", mwStack(http.HandlerFunc(todoHandler.deleteItem)))
	router.Handle("POST /todo/item/restore/{id}", mwStack(http.HandlerFunc(todoHandler.restoreItem)))

	router.Handle("GET /todo/item/create/panel", mwStack(http.HandlerFunc(todoHandler.getCreatePanel)))

	router.Handle("GET /todo/item/{id}/view/panel", mwStack(http.HandlerFunc(todoHandler.getViewPanel)))
	router.Handle("GET /todo/item/{id}/edit/panel", mwStack(http.HandlerFunc(todoHandler.getEditPanel)))

	router.Handle("GET /task/{id}/tag-select", mwStack(http.HandlerFunc(todoHandler.getTagSelector)))
}

func (h *Handler) createItem(w http.ResponseWriter, r *http.Request) {
	currentUserID, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	form := fm.NewTaskForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Task Name")
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description")
	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		form.Errors["Project"] = err.Error()
	}
	projectID, err := strconv.ParseInt(project, 10, 64)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}
	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	dueDate, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}

	if dueDate.Before(time.Now().Add(-24 * time.Hour)) {
		form.Errors["Due Date"] = "Due date cannot be in the past"
	}
	projects, err := h.projectService.GetAllProjectsForUser(currentUserID)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	taskTagIDStrings, err := models.GetListPropertyFromRequest(r, "tags")
	if err != nil {
		form.Errors["Tags"] = err.Error()
	}
	taskTagIDs := make([]int64, len(taskTagIDStrings))
	for i, tagIDString := range taskTagIDStrings {
		tagID, err := strconv.Atoi(tagIDString)
		if err != nil {
			form.Errors["Tags"] = err.Error()
			break
		}
		taskTagIDs[i] = int64(tagID)
	}
	tagsForProject, err := h.projectService.GetProjectTags(currentUserID, projectID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to get project tags for project %d", projectID)
		return
	}

	taskTags := make([]models.TagView, len(taskTagIDStrings))
	for _, tag := range tagsForProject {
		if slices.Contains(taskTagIDs, tag.Id) {
			taskTags = append(taskTags, tag)
		}
	}

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     &dueDate,
		ProjectID:   projectID,
		Tags:        taskTags,
	}
	if len(form.Errors) > 0 {
		if err := templ_shared.EditPanel("Create New Task", templ_todo.TaskFormContent(form, models.ProjectListToMap(projects), tagsForProject)).Render(r.Context(), w); err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to update task form")
		}
		return
	}

	newID, err := h.taskService.CreateTask(currentUserID, projectID, name, description, &dueDate, taskTagIDs)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to create task")
		return
	}

	taskView, err := h.taskService.GetTaskView(newID, currentUserID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve newly created task")
		return
	}

	newForm := fm.NewDefaultTaskForm()
	newForm.Task.ProjectID = taskView.ProjectID

	err = templ_shared.RenderTempls(
		templ_todo.TaskItemCardOOB(taskView),
		templ_shared.EditPanel("Create New Task", templ_todo.TaskFormCard(newForm, models.ProjectListToMap(projects), tagsForProject)),
		templ_shared.ToastMessage("Task created successfully!", "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display new task")
		return
	}
}

type NoItemRowData struct {
	HideNoData bool
}

func (h *Handler) updateItem(w http.ResponseWriter, r *http.Request) {
	currentUserID, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}
	form := fm.NewTaskForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Task Name")
	if err != nil {
		form.Errors["name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description")

	if description == "" {
		form.Errors["Description"] = "Description is required"
	}

	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		form.Errors["Project"] = err.Error()
	}
	projectID, err := strconv.ParseInt(project, 10, 64)
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

	taskTagIDStrings, err := models.GetListPropertyFromRequest(r, "tags")
	if err != nil {
		form.Errors["Tags"] = err.Error()
	}
	taskTagIDs := make([]int64, len(taskTagIDStrings))
	for i, tagIDString := range taskTagIDStrings {
		tagID, err := strconv.Atoi(tagIDString)
		if err != nil {
			form.Errors["Tags"] = err.Error()
			break
		}
		taskTagIDs[i] = int64(tagID)
	}
	tagsForProject, err := h.projectService.GetProjectTags(currentUserID, projectID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to get project tags for project %d", projectID)
		return
	}

	taskTags := make([]models.TagView, len(taskTagIDStrings))
	for _, tag := range tagsForProject {
		if slices.Contains(taskTagIDs, tag.Id) {
			taskTags = append(taskTags, tag)
		}
	}

	form.Task = models.TaskView{
		Name:        name,
		Description: description,
		DueDate:     &date,
		Tags:        taskTags,
	}
	projects, err := h.projectService.GetAllProjectsForUser(currentUserID)
	if err != nil {
		form.Errors["Project"] = err.Error()
	}

	if len(form.Errors) > 0 {
		task, err := h.taskService.GetTaskView(id, currentUserID)
		if err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve task for editing")
			return
		}

		projectTags, err := h.projectService.GetProjectTags(currentUserID, task.ProjectID)
		if err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve project tags")
			return
		}

		if err := templ_shared.EditPanel("Edit Task", templ_todo.TaskItemFormWithErrors(task, models.ProjectListToMap(projects), projectTags, form.Errors)).Render(r.Context(), w); err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to display task edit form")
		}
		return
	}

	err = h.taskService.UpdateTask(
		currentUserID,
		id,
		projectID,
		name,
		description,
		&date,
		currentUserID,
		taskTagIDs,
	)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to update task")
		return
	}
	task, err := h.taskService.GetTaskView(id, currentUserID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve updated task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_todo.TaskCardFrontOOB(task),
		templ_shared.EditPanel("View Task", templ_todo.TaskItemContent(task)),
		templ_shared.ToastMessage("Task updated successfully!", "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task update result")
		return
	}
}

func (h *Handler) getItem(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}
	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code() == errors.ErrNotFound {
			errors.NotFound(h.logger, w, r, err, "Task not found")
			return
		}
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve task")
		return
	}

	var template templ.Component
	source := r.URL.Query().Get("source")
	if source == "list" {
		template = templ_todo.TaskItemContentOOBTargetList(task)
	} else {
		contentType := r.Header.Get("accept")
		if contentType == "text/html" {
			template = templ_todo.TaskItem(task, models.ScreenTodo)
		} else {
			template = templ_todo.TaskItemWithBody(task, models.ScreenTodo)
		}
	}
	if err := template.Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task")
		return
	}
}

func (h *Handler) updateItemStatus(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}

	err = h.taskService.UpdateTaskStatus(currentUserid, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to update task status")
		return
	}

	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve updated task")
		return
	}
	source := r.URL.Query().Get("source")
	if source == "task_page" {
		if err = templ_shared.RenderTempls(
			templ_todo.TaskItemContent(task),
			templ_todo.TaskCardFrontOOB(task),
		).Render(r.Context(), w); err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to display updated task")
			return
		}
	} else {
		if err = templ_todo.TaskCardFront(task).Render(r.Context(), w); err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to display updated task")
			return
		}
	}
}

func (h *Handler) listItems(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	projects, err := h.projectService.GetAllProjectsForUser(currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "failed to retrieve projects")
		return
	}

	tmpl := templ_todo.TaskListWithBody(models.ScreenTodo, fm.NewTaskSearchForm(), models.ProjectListToMap(projects))
	if err := tmpl.Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task list")
		return
	}
}

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	currentUserid, currentUseremail, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}
	searchText := models.GetPropertyFromRequest(r, "search")
	tags := models.GetPropertyFromRequest(r, "tags")

	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Project select value malformed")
	}
	projectID, err := strconv.ParseInt(project, 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Unable to read project id")
	}
	status := models.GetPropertyFromRequest(r, "project")
	h.logger.Debug("Got search fields", "search_text", searchText, "tags", tags, "project", project, "status", status)
	tasks, err := h.taskService.Search(currentUserid, searchText, projectID)
	if err != nil {
		h.logger.Error("failed to search for tasks", "error", err, "user_email", currentUseremail)
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve tasks")
		return
	}

	if err := templ_todo.TaskListContentOOB(tasks).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task list")
		return
	}
}

func (h *Handler) listItemsLazy(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	tasks, err := h.taskService.GetTaskViewList(currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve tasks")
		return
	}
	if len(tasks) == 0 {
		logger.TODO("Implement a no-data-row for tasks")
	}

	err = templ_todo.TaskListContent(tasks).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task list content")
		return
	}
}

func (h *Handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}

	err = h.taskService.DeleteTask(currentUserid, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to delete task")
		return
	}

	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve deleted task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_shared.ToastMessage("Task Deleted", "warning"),
		templ_todo.TaskCardFront(task),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task deletion result")
		return
	}
}

func (h *Handler) restoreItem(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}

	err = h.taskService.RestoreTask(currentUserid, id)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to restore task")
		return
	}

	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve restored task")
		return
	}

	err = templ_shared.RenderTempls(
		templ_shared.ToastMessage("Task Restored", "success"),
		templ_todo.TaskCardFront(task),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to display task restoration result")
		return
	}
}

func (h *Handler) getCreatePanel(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	projects, err := h.projectService.GetAllProjectsForUser(currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve projects")
		return
	}

	form := fm.NewDefaultTaskForm()

	projectID := r.URL.Query().Get("project_id")
	if projectID != "" {
		if id, err := strconv.ParseInt(projectID, 10, 64); err == nil {
			form.SetProject(id)
		}
	} else {
		form.Task.ProjectID = projects[0].Id
	}
	tagsForProject := []models.TagView{}
	if form.Task.ProjectID != 0 {
		tagsForProject, err = h.projectService.GetProjectTags(currentUserid, form.Task.ProjectID)
		if err != nil {
			errors.InternalServerError(h.logger, w, r, err, "Failed to get project tags for project %d", form.Task.ProjectID)
			return
		}
	}

	if err := templ_shared.EditPanel("Create New Task", templ_todo.TaskFormCard(form, models.ProjectListToMap(projects), tagsForProject)).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to render task create panel")
		return
	}
}

func (h *Handler) getEditPanel(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}

	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve task")
		return
	}

	projects, err := h.projectService.GetAllProjectsForUser(currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve projects")
		return
	}

	projectTags, err := h.projectService.GetProjectTags(currentUserid, task.ProjectID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve project tags")
		return
	}

	if err := templ_shared.EditPanel("Edit Task", templ_todo.TaskItemForm(task, models.ProjectListToMap(projects), projectTags)).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to render task edit panel")
		return
	}
}

func (h *Handler) getViewPanel(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Invalid task ID provided")
		return
	}

	task, err := h.taskService.GetTaskView(id, currentUserid)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve task")
		return
	}

	if err := templ_shared.EditPanel("View Task", templ_todo.TaskItemContent(task)).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to render task edit panel")
		return
	}
}

func formDataFromTask(task *models.TaskView) fm.TaskForm {
	formData := fm.NewTaskForm()
	formData.Task.Name = task.Name
	formData.Task.Description = task.Description
	formData.Task.AssignedTo = task.AssignedTo
	formData.Task.DueDate = task.DueDate
	formData.Task.ProjectID = task.ProjectID
	formData.Task.ProjectName = task.ProjectName

	return formData
}

func (h *Handler) getTagSelector(w http.ResponseWriter, r *http.Request) {
	currentUserid, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(h.logger, w, r, err, "user auth failed")
		return
	}

	project, err := models.GetRequiredPropertyFromRequest(r, "project", "Project")
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Failed to read selected project")
		return
	}
	projectID, err := strconv.ParseInt(project, 10, 64)
	if err != nil {
		errors.BadRequest(h.logger, w, r, err, "Failed to read selected project")
		return
	}

	tags, err := h.projectService.GetProjectTags(currentUserid, projectID)
	if err != nil {
		errors.InternalServerError(h.logger, w, r, err, "Failed to retrieve project tags")
		return
	}
	if err = templ_todo.TagSelect([]models.TagView{}, tags).Render(r.Context(), w); err != nil {
		errors.InternalServerError(h.logger, w, r, err, "failed to render project content for id %d", projectID)
		return
	}
}
