package project

import (
	"fmt"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/errors"
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

var (
	handlerSource = "ProjectHandler"
	defaultForm   = form_models.NewDefaultProjectForm()
)

func SetupProjectHandler(service project_service.ProjectService, task_service task_service.TaskService, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(handlerSource)

	projectsHandler := &Handler{
		project_service: service,
		task_service:    task_service,
		logger:          logger,
	}

	router.Handle("GET /projects", mw_stack(http.HandlerFunc(projectsHandler.getProjects)))
	router.Handle("POST /project", mw_stack(http.HandlerFunc(projectsHandler.createProject)))
	router.Handle("PUT /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.updateProject)))

	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
	router.Handle("DELETE /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.deleteProject)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.FrontendErrorBadRequest(w, r, h.logger, err, "id for project is not a valid int")
		return
	}

	// NOTE: service call
	projectView, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		errors.FrontendErrorNotFound(w, r, h.logger, err, "failed to find project %d", id)
		return
	}

	// NOTE: Check if this is an edit panel request
	source := r.URL.Query().Get("source")
	if source == "list" {
		if err = templ_project.ProjectContentOOB(*projectView, false, map[string]string{}).Render(r.Context(), w); err != nil {
			errors.FrontendError(w, r, h.logger, err, "failed to render project content for id %d", id)
			return
		}
		return
	}

	// NOTE: frontend response for full page
	formData := formDataFromProject(*projectView)

	var tasks []*models.TaskView
	tasks, err = h.task_service.GetProjectTaskViewList(current_user_id, projectView.Id)
	if err != nil {
		h.logger.Error(err, "failed to get tasks for project %d", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templ_project.ProjectWithBody(*projectView, models.ScreenProjects, formData, form_models.NewDefaultTaskForm(), tasks).Render(r.Context(), w)
	if err != nil {
		h.logger.Error(err, "failed to render project page for id %d", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, _, err := helpers.GetUserFromContext(r)
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

	// NOTE: service call
	hasProjects, err := h.project_service.DeleteProject(id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to delete project")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// NOTE: frontend response
	project_view, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to find project")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templ_project.ProjectCardFront(*project_view).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		h.logger.Error(err, "failed to render task row")
		return
	}
	if err := templ_shared.NoDataRowOOB(hasProjects).Render(r.Context(), w); err != nil {
		// if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		h.logger.Error(err, "failed to set no-data status")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: service call
	pList, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to get projects for user %d", current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// NOTE: frontend response
	form := form_models.NewDefaultProjectForm()
	err = templ_project.ProjectListWithBody(models.ScreenProjects, form, pList).Render(r.Context(), w)
	if err != nil {
		h.logger.Error(err, "failed to render project list page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

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
		DueDate:     &due_date,
	}

	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		if err := templ_project.ProjectFormContent(form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, handlerSource, "Failed to render template for formData")
		}
		return
	}

	now := time.Now()
	new_id, err := h.project_service.CreateProject(
		current_user_id,
		current_user_id,
		name, description,
		&now, &due_date)
	if err != nil {
		h.logger.Error(err, "failed to create project for user %d", current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	projectView, err := h.project_service.GetProjectView(new_id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to get newly added project (%d) for user %d", new_id, current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := templ_project.ProjectCardOOB(*projectView).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render new project row (%d) for user %d", new_id, current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render no data row for user %d", current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templ_project.ProjectFormContent(defaultForm).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render project form for user %d", current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templ_shared.ToastActionRedirect("Successfully created project", fmt.Sprintf("/project/%d", new_id), "Go to project", "success").Render(r.Context(), w)
	if err != nil {
		h.logger.Error(err, "failed to render toast for user %d", current_user_id)
		return
	}
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	current_user_id, _, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.FrontendErrorUnauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		h.logger.Error(err, "not authenticated")
		return
	}
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
		DueDate:     &due_date,
	}

	if len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		projectView, err := h.project_service.GetProjectView(id, current_user_id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := templ_project.ProjectContentOOB(*projectView, true, form.Errors).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, handlerSource, "Failed to render template for formData")
		}
		return
	}

	err = h.project_service.UpdateProject(
		id, current_user_id, current_user_id, // NOTE: for now owner is also creator
		name, description,
		&due_date)
	if err != nil {
		errors.FrontendError(w, r, h.logger, err, "failed to update project %d", id)
		return
	}
	projectView, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		h.logger.Error(err, "failed to get updated added project (%d) for user %d", id, current_user_id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = templ_project.ProjectContentOOB(*projectView, false, nil).Render(r.Context(), w)
	if err != nil {
		// TODO: Handle error on frontend
		h.logger.Error(err, "Failed to execute template for the project page")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := templ_project.ProjectCardFrontOOB(*projectView).Render(r.Context(), w); err != nil {
		h.logger.Error(err, "failed to render project card OOB")
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
