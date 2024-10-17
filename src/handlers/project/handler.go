package project

import (
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
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

	router.Handle("GET /project/{id}", mw_stack(http.HandlerFunc(projectsHandler.getProject)))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	assert.IsTrue(false, h.logger, "Not implemented")
}
