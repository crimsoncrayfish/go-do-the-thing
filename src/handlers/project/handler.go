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
	"net/http"
)

type Handler struct {
	logger           slog.Logger
	ProjectRepo      projects_repo.ProjectsRepo
	ProjectUsersRepo project_users_repo.ProjectUsersRepo
	RolesRepo        roles_repo.RolesRepo
}

var activeScreens models.NavBarObject

func SetupProjectHandler(projectRepo projects_repo.ProjectsRepo, projectUsersRepo project_users_repo.ProjectUsersRepo, rolesRepo roles_repo.RolesRepo, router *http.ServeMux, mw_stack middleware.Middleware) {
	logger := slog.NewLogger("ProjectsHandler")

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
	assert.IsTrue(false, "Not implemented")
}

func (h *Handler) getProjects(w http.ResponseWriter, r *http.Request) {
	// NOTE: Auth check
	_, _, _, err := helpers.GetUserFromContext(r)
	assert.NoError(err, h.logger, "user auth failed unsuccessfully")

	pl := make([]models.Project, 0)

	if err := templ_project.ProjectListWithBody(activeScreens, pl).Render(r.Context(), w); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		assert.NoError(err, h.logger, "Failed to render template for formData")
	}
	return
}
