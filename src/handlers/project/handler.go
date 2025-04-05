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
	form_models "go-do-the-thing/src/models/forms"
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

var (
	source      = "ProjectsHandler"
	defaultForm = form_models.NewDefaultProjectForm()
)

func SetupProjectHandler(projectRepo projects_repo.ProjectsRepo, projectUsersRepo project_users_repo.ProjectUsersRepo, rolesRepo roles_repo.RolesRepo, usersRepo users_repo.UsersRepo, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger(source)

	activeScreens = models.NavBarObject{ActiveScreens: models.ActiveScreens{IsProjects: true}}
	projectsHandler := &Handler{
		ProjectRepo:      projectRepo,
		ProjectUsersRepo: projectUsersRepo,
		UsersRepo:        usersRepo,
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
	id, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	pl, err := h.ProjectRepo.GetProjects(id)
	assert.NoError(err, source, "failed to get projects")

	form := form_models.NewDefaultProjectForm()

	pl_v, err := h.projectsToViewModels(pl)
	assert.NoError(err, source, "failed to convert the project list")
	if err := templ_project.ProjectListWithBody(activeScreens, form, pl_v).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		assert.NoError(err, source, "Failed to render template for formData")
	}
}

func (h *Handler) createProjectUI(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, currentUserEmail, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	form := form_models.NewProjectForm()
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

	// TODO: Capture startdate
	startDate := database.SqLiteNow()
	project := models.Project{
		Name:         name,
		Description:  description,
		Owner:        currentUserId,
		StartDate:    startDate,
		DueDate:      database.NewSqliteTime(date),
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsComplete:   false,
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
	err = h.ProjectUsersRepo.Insert(id, currentUserId, 0)
	if err != nil {
		h.logger.Error(err, "failed to insert project user")
		form.Errors["ProjectUser"] = "failed to create project user link"
		if err := templ_project.ProjectFormContent("Create", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for project")
			// TODO: what should happen if the fetch fails after create
		}
		assert.IsTrue(false, "InsertProject", "There is now an unlinked project with id %d intended for user %s", id, currentUserEmail)
		return
	}
	project, err = h.ProjectRepo.GetProject(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to get newly inserted project")
		// TODO: what should happen if the fetch fails after create
		return
	}

	// NOTE: Success zone

	owner, err := h.UsersRepo.GetUserById(project.ModifiedBy)
	if ok := h.handleUserIdNotFound(err, project.ModifiedBy); !ok {
		assert.NoError(err, source, "how does a project with an owner user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}
	createdByUser, err := h.UsersRepo.GetUserById(project.CreatedBy)
	if ok := h.handleUserIdNotFound(err, project.CreatedBy); !ok {
		assert.NoError(err, source, "how does a project with an created by user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}
	modifiedByUser, err := h.UsersRepo.GetUserById(project.ModifiedBy)
	if ok := h.handleUserIdNotFound(err, project.ModifiedBy); !ok {
		assert.NoError(err, source, "how does a project with an modified by user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}

	projectListItem := project.ToViewModel(owner, createdByUser, modifiedByUser)
	if err := templ_project.ProjectRowOOB(projectListItem).Render(r.Context(), w); err != nil {
		assert.NoError(err, source, "failed to render new project row with id %d", project.Id)
		// TODO: what should happen if the fetch fails after create
		return
	}
	if err := templ_shared.NoDataRowOOB(true).Render(r.Context(), w); err != nil {
		// if err = h.templates.RenderOk(w, "no-data-row-oob", to); err != nil {
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

func formDataFromProject(project models.Project, assignedUser string) form_models.ProjectForm {
	formData := form_models.NewProjectForm()
	formData.Project.Name = project.Name
	formData.Project.Description = project.Description
	formData.Project.DueDate = project.DueDate
	return formData
}

func (h *Handler) projectsToViewModels(projects []models.Project) ([]models.ProjectView, error) {
	projectViews := make([]models.ProjectView, len(projects))

	for i, project := range projects {
		owner, err := h.UsersRepo.GetUserById(project.Owner)
		assert.NoError(err, source, "Failed to get owner user")

		// Fetch the CreatedBy user
		var createdBy models.User
		if project.Owner == project.CreatedBy {
			createdBy = owner
		} else {
			createdBy, err = h.UsersRepo.GetUserById(project.CreatedBy)
			assert.NoError(err, source, "Failed to get created by user")
		}

		var modifiedBy models.User
		switch project.ModifiedBy {
		case project.Owner:
			modifiedBy = owner
		case project.CreatedBy:
			modifiedBy = createdBy
		default:
			modifiedBy, err = h.UsersRepo.GetUserById(project.ModifiedBy)
			assert.NoError(err, source, "failed to get modified by user")
		}

		// Convert to ViewModel
		projectView := project.ToViewModel(owner, createdBy, modifiedBy)

		// Fetch owner if ID does not equal zero.
		projectViews[i] = projectView
	}

	return projectViews, nil
}
