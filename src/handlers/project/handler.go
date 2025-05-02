package project

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

	templ_project "go-do-the-thing/src/handlers/project/templ"

	form_models "go-do-the-thing/src/models/forms"
	project_service "go-do-the-thing/src/services/project"
	task_service "go-do-the-thing/src/services/task"
	templ_shared "go-do-the-thing/src/shared/templ"
)

type Handler struct {
	logger          slog.Logger
	project_service project_service.ProjectService
	task_service    task_service.TaskService
}

var activeScreens models.NavBarObject

var (
	handlerSource = "ProjectHandler"
	defaultForm   = form_models.NewDefaultProjectForm()
)

func SetupProjectHandler(service project_service.ProjectService, task_service task_service.TaskService, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(handlerSource)

	activeScreens = models.NavBarObject{ActiveScreens: models.ActiveScreens{IsProjects: true}}
	projectsHandler := &Handler{
		project_service: service,
		task_service:    task_service,
		logger:          logger,
	}

	router.Handle("GET /projects", mw_stack(http.HandlerFunc(projectsHandler.getProjects)))
	router.Handle("POST /project", mw_stack(http.HandlerFunc(projectsHandler.createProject)))

	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
	router.Handle("DELETE /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.deleteProject)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, handlerSource, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: service call
	projectView, err := h.project_service.GetProjectView(id, currentUserId)
	if err != nil {
		h.logger.Error(err, "Failed to get the project")
		// TODO: Handle error on frontend
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// NOTE: frontend response
	contentType := r.Header.Get("accept")
	formData := formDataFromProject(*projectView)

	var tasks []*models.TaskView
	tasks, err = h.task_service.GetProjectTaskViewList(currentUserId, projectView.Id)
	if err != nil {
		h.logger.Error(err, "Failed to get the tasks")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch contentType {
	case "text/html":
		err = templ_project.ProjectView(*projectView, activeScreens, formData, tasks).Render(r.Context(), w)
	default:
		err = templ_project.ProjectWithBody(*projectView, activeScreens, formData, tasks).Render(r.Context(), w)
	}

	if err != nil {
		// TODO: Handle error on frontend
		h.logger.Error(err, "Failed to execute template for the project page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, handlerSource, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: service call
	hasProjects, err := h.project_service.DeleteProject(id, currentUserId)
	if err != nil {
		h.logger.Error(err, "failed to delete project")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: frontend response
	if err := templ_shared.NoDataRowOOB(hasProjects).Render(r.Context(), w); err != nil {
		// if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		h.logger.Error(err, "failed to set no-data status")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, handlerSource, "user auth failed unsuccessfully")

	// NOTE: service call
	pList, err := h.project_service.GetAllProjectsForUser(currentUserId)
	if err != nil {
		h.logger.Error(err, "failed to get projects for user %d", currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// NOTE: frontend response
	form := form_models.NewDefaultProjectForm()
	err = templ_project.ProjectListWithBody(activeScreens, form, pList).Render(r.Context(), w)
	if err != nil {
		h.logger.Error(err, "failed to get projects for user %d", currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, handlerSource, "user auth failed unsuccessfully")

	// NOTE: Collect data
	form := form_models.NewProjectForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Project Name")
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description", "Description")

	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	due_date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}

	form.Project = models.ProjectView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(due_date),
	}

	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := templ_project.ProjectFormContent("Create", form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, handlerSource, "Failed to render template for formData")
		}
		return
	}

	new_id, err := h.project_service.CreateProject(
		currentUserId, currentUserId, // NOTE: for now owner is also creator
		name, description,
		database.SqLiteNow(), database.NewSqliteTime(due_date))
	if err != nil {
		h.logger.Error(err, "failed to create project for user %d", currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	projectView, err := h.project_service.GetProjectView(new_id, currentUserId)
	if err != nil {
		h.logger.Error(err, "failed to get newly added project (%d) for user %d", new_id, currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templ_project.ProjectRowOOB(*projectView).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render new project row (%d) for user %d", new_id, currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render no data row for user %d", currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templ_project.ProjectFormContent("Create", defaultForm).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render project form for user %d", currentUserId)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func formDataFromProject(project models.ProjectView) form_models.ProjectForm {
	formData := form_models.NewProjectForm()
	formData.Project.Name = project.Name
	formData.Project.Description = project.Description
	formData.Project.DueDate = project.DueDate
	return formData
}
