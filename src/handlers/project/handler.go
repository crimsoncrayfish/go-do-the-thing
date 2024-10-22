package project

import (
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
	templ_project "go-do-the-thing/src/handlers/project/templ"
	"go-do-the-thing/src/helpers"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/middleware"
	"go-do-the-thing/src/models"
	"go-do-the-thing/src/models/forms"
	"net/http"
)

type Handler struct {
	logger           slog.Logger
	ProjectRepo      projects_repo.ProjectsRepo
	ProjectUsersRepo project_users_repo.ProjectUsersRepo
	RolesRepo        roles_repo.RolesRepo
}

var activeScreens models.NavBarObject

var source = assert.Source{"ProjectsHandler"}

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
