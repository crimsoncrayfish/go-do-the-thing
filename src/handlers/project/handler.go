package project

import (
	"database/sql"
	"errors"
	"go-do-the-thing/src/database"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
	users_repo "go-do-the-thing/src/database/repos/users"
	templ_project "go-do-the-thing/src/handlers/project/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"go-do-the-thing/src/models/forms"
	fm "go-do-the-thing/src/models/forms"
	templ_shared "go-do-the-thing/src/shared/templ"
	"net/http"
	"time"
)

type Handler struct {
	logger           slog.Logger
	ProjectRepo      projects_repo.ProjectsRepo
	UsersRepo        users_repo.UsersRepo
	ProjectUsersRepo project_users_repo.ProjectUsersRepo
	RolesRepo        roles_repo.RolesRepo
}

var activeScreens models.NavBarObject

var source = assert.Source{Name: "ProjectsHandler"}
var defaultForm = fm.NewDefaultProjectForm()

func SetupProjectHandler(projectRepo projects_repo.ProjectsRepo, projectUsersRepo project_users_repo.ProjectUsersRepo, rolesRepo roles_repo.RolesRepo, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source.Name)

	activeScreens = models.NavBarObject{ActiveScreens: models.ActiveScreens{IsProjects: true}}
	projectsHandler := &Handler{
		ProjectRepo:      projectRepo,
		ProjectUsersRepo: projectUsersRepo,
		RolesRepo:        rolesRepo,
		logger:           logger,
	}

	router.Handle("GET /projects", mw_stack(http.HandlerFunc(projectsHandler.getProjects)))
	router.Handle("POST /project", mw_stack(http.HandlerFunc(projectsHandler.createProjectUI)))
	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	assert.IsTrue(false, source, "Not implemented")
}

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	pl := make([]models.ProjectView, 0)

	form := form_models.NewDefaultProjectForm()

	if err := templ_project.ProjectListWithBody(activeScreens, form, pl).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		assert.NoError(err, source, "Failed to render template for formData")
	}
	return
}
func (h *Handler) createProjectUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	form := fm.NewProjectForm()
	name, err := models.GetPropertyFromRequest(r, "name", "Project Name", true)
	if err != nil {
		form.Errors["Name"] = err.Error()
	}
	description, _ := models.GetPropertyFromRequest(r, "description", "Description", false)

	dateRaw, err := models.GetPropertyFromRequest(r, "due_date", "Due on", true)
	if err != nil {
		form.Errors["Due Date"] = err.Error()
	}
	date, err := time.Parse("2006-01-02", dateRaw)

	form.Project = models.ProjectView{
		Name:        name,
		Description: description,
		DueDate:     database.NewSqliteTime(date),
	}
	if err != nil || len(form.Errors) > 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := templ_project.ProjectFormContent("Create", form).Render(r.Context(), w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			assert.NoError(err, source, "Failed to render template for formData")
		}
		return
	}

	project := models.Project{
		Name:         name,
		Description:  description,
		DueDate:      database.NewSqliteTime(date),
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
	}

	// NOTE: Validate data
	form = formDataFromProject(project, currentUserEmail)

	// NOTE: Take action
	id, err := h.ProjectRepo.InsertProject(currentUserId, project)
	if err != nil {
		h.logger.Error(err, "failed to insert project")
		form.Errors["Project"] = "failed to create project"
		if err := templ_project.ProjectFormContent("Create", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for project")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	project, err = h.ProjectRepo.GetProject(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to get newly inserted project")
		// TODO: what should happen if the fetch fails after create
		return
	}

	// NOTE: Success zone

	createdByUser, err := h.UsersRepo.GetUserById(project.CreatedBy)
	if ok := h.handleUserIdNotFound(err, project.CreatedBy); !ok {
		assert.NoError(err, source, "how does a project with an created by user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}

	projectListItem := models.ProjectToViewModel(project, createdByUser)
	if err := templ_project.ProjectRowOOB(projectListItem).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render new project row with id %d", project.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		//if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
		assert.NoError(err, source, "failed to render no data row")
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_project.ProjectFormContent("Create", defaultForm).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render the project form after creation")
		// TODO: what should happen if rendering fails
		return
	}
}
func (h *Handler) handleUserIdNotFound(err error, userId int64) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, sql.ErrNoRows) {
		h.logger.Error(err, "the entered email address does not corrispond to an existing user: %d", userId)
	} else {
		assert.NoError(err, source, "some error occurred. probably fialed to read from the db while checking user %d", userId)
	}
	return false
}

func formDataFromProject(project models.Project, assignedUser string) fm.ProjectForm {
	formData := fm.NewProjectForm()
	formData.Project.Name = project.Name
	formData.Project.Description = project.Description
	formData.Project.DueDate = project.DueDate
	return formData
}
