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
	"strconv"
	"time"
)

type Handler struct {
	logger           slog.Logger
	projectRepo      projects_repo.ProjectsRepo
	usersRepo        users_repo.UsersRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
	rolesRepo        roles_repo.RolesRepo
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
		projectRepo:      projectRepo,
		projectUsersRepo: projectUsersRepo,
		usersRepo:        usersRepo,
		rolesRepo:        rolesRepo,
		logger:           logger,
	}

	router.Handle("GET /projects", mw_stack(http.HandlerFunc(projectsHandler.getProjects)))
	router.Handle("POST /project", mw_stack(http.HandlerFunc(projectsHandler.createProject)))

	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
	router.Handle("DELETE /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.deleteProject)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetProject(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to delete project")
		return
	}
	// NOTE: Take action
	owner, err := h.usersRepo.GetUserById(project.Owner)
	assert.NoError(err, source, "this user should exist since they own project: %d", project.Owner)
	formData := formDataFromProject(project, owner.Email)

	// NOTE: Success zone
	var createdBy models.User
	// TODO: Create a layer between the repos and the handler to enrich an object and handle all this validation
	if project.CreatedBy == project.Owner {
		createdBy = owner
	} else {
		createdBy, err = h.usersRepo.GetUserById(project.CreatedBy)
		if ok := h.handleUserIdNotFound(err, project.CreatedBy); !ok {
			assert.NoError(err, source, "how does a project with an created by user id of %d even exist?", project.CreatedBy)
			return
		}
	}
	var modifiedBy models.User
	// TODO: Create a layer between the repos and the handler to enrich an object and handle all this validation
	switch project.ModifiedBy {
	case project.Owner:
		modifiedBy = owner
	case project.CreatedBy:
		modifiedBy = createdBy
	default:
		modifiedBy, err = h.usersRepo.GetUserById(project.ModifiedBy)
		if ok := h.handleUserIdNotFound(err, project.ModifiedBy); !ok {
			assert.NoError(err, source, "how does a project with an modified by user id of %d even exist?", project.ModifiedBy)
			return
		}
	}

	projectView := project.ToViewModel(owner, createdBy, modifiedBy)
	contentType := r.Header.Get("accept")
	if contentType == "text/html" {
		if err = templ_project.ProjectView(projectView, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the project page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if err = templ_project.ProjectWithBody(projectView, activeScreens, formData).Render(r.Context(), w); err != nil {
			// TODO: some user feedback here?
			h.logger.Error(err, "Failed to execute template for the project page")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	currentUserId, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	// NOTE: Collect data
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.logger.Error(err, "failed to parse id from path")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// NOTE: Take action
	err = h.projectRepo.DeleteProject(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to delete project")
		return
	}

	// NOTE: Success zone
	hasData, err := h.projectRepo.GetProjectCount(currentUserId)
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

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	id, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, source, "user auth failed unsuccessfully")

	pl, err := h.projectRepo.GetProjects(id)
	assert.NoError(err, source, "failed to get projects")

	form := form_models.NewDefaultProjectForm()

	pl_v, err := h.projectsToViewModels(pl)
	assert.NoError(err, source, "failed to convert the project list")
	if err := templ_project.ProjectListWithBody(activeScreens, form, pl_v).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		assert.NoError(err, source, "Failed to render template for formData")
	}
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
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
	id, err := h.projectRepo.InsertProject(currentUserId, project)
	if err != nil {
		h.logger.Error(err, "failed to insert project")
		form.Errors["Project"] = "failed to create project"
		if err := templ_project.ProjectFormContent("Create", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for project")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	err = h.projectUsersRepo.Insert(id, currentUserId, 0)
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
	project, err = h.projectRepo.GetProject(id, currentUserId)
	if err != nil {
		assert.NoError(err, source, "failed to get newly inserted project")
		// TODO: what should happen if the fetch fails after create
		return
	}

	// NOTE: Success zone

	owner, err := h.usersRepo.GetUserById(project.ModifiedBy)
	if ok := h.handleUserIdNotFound(err, project.ModifiedBy); !ok {
		assert.NoError(err, source, "how does a project with an owner user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}
	createdByUser, err := h.usersRepo.GetUserById(project.CreatedBy)
	if ok := h.handleUserIdNotFound(err, project.CreatedBy); !ok {
		assert.NoError(err, source, "how does a project with an created by user id of %d even exist?", project.CreatedBy)
		// TODO: what should happen if the fetch fails after create
		return
	}
	modifiedByUser, err := h.usersRepo.GetUserById(project.ModifiedBy)
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
		owner, err := h.usersRepo.GetUserById(project.Owner)
		assert.NoError(err, source, "Failed to get owner user")

		// Fetch the CreatedBy user
		var createdBy models.User
		if project.Owner == project.CreatedBy {
			createdBy = owner
		} else {
			createdBy, err = h.usersRepo.GetUserById(project.CreatedBy)
			assert.NoError(err, source, "Failed to get created by user")
		}

		var modifiedBy models.User
		switch project.ModifiedBy {
		case project.Owner:
			modifiedBy = owner
		case project.CreatedBy:
			modifiedBy = createdBy
		default:
			modifiedBy, err = h.usersRepo.GetUserById(project.ModifiedBy)
			assert.NoError(err, source, "failed to get modified by user")
		}

		// Convert to ViewModel
		projectView := project.ToViewModel(owner, createdBy, modifiedBy)

		// Fetch owner if ID does not equal zero.
		projectViews[i] = projectView
	}

	return projectViews, nil
}
