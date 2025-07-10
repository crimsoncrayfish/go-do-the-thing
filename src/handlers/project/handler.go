package project

import (
	"go-do-the-thing/src/helpers"
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
	router.Handle("GET /projects/lazy", mw_stack(http.HandlerFunc(projectsHandler.getProjectsLazy)))
	router.Handle("POST /project", mw_stack(http.HandlerFunc(projectsHandler.createProject)))
	router.Handle("PUT /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.updateProject)))

	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
	router.Handle("DELETE /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.deleteProject)))

	router.Handle("GET /project/create/panel", mw_stack(http.HandlerFunc(projectsHandler.getCreatePanel)))
	router.Handle("GET /project/{id}/edit/panel", mw_stack(http.HandlerFunc(projectsHandler.getEditPanel)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid project ID provided")
		return
	}

	projectView, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok && appErr.Code() == errors.ErrNotFound {
			errors.NotFound(w, r, h.logger, err, "Project not found")
			return
		}
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve project")
		return
	}

	source := r.URL.Query().Get("source")
	if source == "list" {
		if err = templ_project.ProjectContentOOB(*projectView, false, map[string]string{}).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, h.logger, err, "failed to render project content for id %d", id)
			return
		}
		return
	}

	formData := formDataFromProject(*projectView)

	var tasks []*models.TaskView
	tasks, err = h.task_service.GetProjectTaskViewList(current_user_id, projectView.Id)
	if err != nil {
		h.logger.Error(err, "failed to get tasks for project %d", id)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = templ_project.ProjectWithBody(*projectView, models.ScreenProjects, formData, form_models.NewDefaultTaskForm(), tasks).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "failed to render project page")
		return
	}
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid project ID provided")
		return
	}

	hasProjects, err := h.project_service.DeleteProject(id, current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to delete project")
		return
	}

	project_view, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve project information")
		return
	}

	err = templ_shared.RenderTempls(
		templ_project.ProjectCardFront(*project_view),
		templ_shared.NoDataRowOOB(hasProjects),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display project deletion result")
		return
	}
}

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	err = templ_project.ProjectListWithBody(models.ScreenProjects).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display project list")
		return
	}
}

func (h *Handler) getProjectsLazy(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	projects, err := h.project_service.GetAllProjectsForUser(current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve projects")
		return
	}

	err = templ_project.ProjectListContent(projects).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display project list content")
		return
	}
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	form := form_models.NewProjectForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Project Name")
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description", "Description")

	if description == "" {
		form.Errors["Description"] = "Description is required"
	}

	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	due_date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}

	if due_date.Before(time.Now()) {
		form.Errors["Due Date"] = "Due date cannot be in the past"
	}

	form.Project = models.ProjectView{
		Name:        name,
		Description: description,
		DueDate:     &due_date,
	}

	if len(form.Errors) > 0 {
		if err := templ_shared.EditPanel("Create New Project", templ_project.ProjectFormCard(form)).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, h.logger, err, "Failed to display project form")
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
		errors.InternalServerError(w, r, h.logger, err, "Failed to create project")
		return
	}
	projectView, err := h.project_service.GetProjectView(new_id, current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve newly created project")
		return
	}

	defaultForm := form_models.NewDefaultProjectForm()

	err = templ_shared.RenderTempls(
		templ_project.ProjectCardOOB(*projectView),
		templ_shared.EditPanel("Create New Project", templ_project.ProjectFormCard(defaultForm)),
		templ_shared.ToastMessage("Project created successfully!", "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display project creation result")
		return
	}
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid project ID provided")
		return
	}
	form := form_models.NewProjectForm()
	name, err := models.GetRequiredPropertyFromRequest(r, "name", "Project Name")
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description := models.GetPropertyFromRequest(r, "description", "Description")

	if description == "" {
		form.Errors["Description"] = "Description is required"
	}

	dateRaw, err := models.GetRequiredPropertyFromRequest(r, "due_date", "Due on")
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	due_date, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}

	if due_date.Before(time.Now()) {
		form.Errors["Due Date"] = "Due date cannot be in the past"
	}

	form.Project = models.ProjectView{
		Name:        name,
		Description: description,
		DueDate:     &due_date,
	}

	if len(form.Errors) > 0 {
		project, err := h.project_service.GetProjectView(id, current_user_id)
		if err != nil {
			errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve project for editing")
			return
		}
		if err := templ_shared.EditPanel("Edit Project", templ_project.ProjectContentOOB(*project, true, form.Errors)).Render(r.Context(), w); err != nil {
			errors.InternalServerError(w, r, h.logger, err, "Failed to display project edit form")
		}
		return
	}

	err = h.project_service.UpdateProject(
		id, current_user_id, current_user_id,
		name, description,
		&due_date)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "failed to update project %d", id)
		return
	}
	projectView, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve updated project")
		return
	}

	err = templ_shared.RenderTempls(
		templ_project.ProjectContentOOB(*projectView, false, map[string]string{}),
		templ_project.ProjectCardFrontOOB(*projectView),
		templ_shared.ToastMessage("Project updated successfully!", "success"),
	).Render(r.Context(), w)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to display project update result")
		return
	}
}

func (h *Handler) getCreatePanel(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	form := form_models.NewDefaultProjectForm()

	if err := templ_shared.EditPanel("Create New Project", templ_project.ProjectFormCard(form)).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render project create panel")
		return
	}
}

func (h *Handler) getEditPanel(w http.ResponseWriter, r *http.Request) {
	current_user_id, _, _, err := helpers.GetUserFromContext(r)
	if err != nil {
		errors.Unauthorized(w, r, h.logger, err, "user auth failed")
		return
	}

	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		errors.BadRequest(w, r, h.logger, err, "Invalid project ID provided")
		return
	}

	project, err := h.project_service.GetProjectView(id, current_user_id)
	if err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to retrieve project")
		return
	}
	project.Tags = []models.TagView{models.NewTag(1, "#33FF57"), models.NewTag(2, "#ff5733"), models.NewTag(3, "#5733FF")}

	form := form_models.NewProjectForm()
	form.Project = *project

	if err := templ_shared.EditPanel("Edit Project", templ_project.ProjectContent(*project, false, map[string]string{})).Render(r.Context(), w); err != nil {
		errors.InternalServerError(w, r, h.logger, err, "Failed to render project edit panel")
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
