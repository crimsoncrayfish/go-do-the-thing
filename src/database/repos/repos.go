package repos

import (
	"go-do-the-thing/src/database"
	project_repos "go-do-the-thing/src/database/repos/projects"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type RepoContainer struct {
	// NOTE: User data tables
	usersRepo *UsersRepo

	// NOTE: Enum-like tables
	tagsRepo  *TagsRepo
	rolesRepo *project_repos.RolesRepo

	// NOTE: Projects
	projectsRepo     *project_repos.ProjectsRepo
	projectUsersRepo *project_repos.ProjectUsersRepo
	projectTagsRepo  *project_repos.ProjectTagsRepo

	// NOTE: Tasks
	tasksRepo    *TasksRepo
	taskTagsRepo *TaskTagsRepo
}

func NewContainer(connection database.DatabaseConnection) *RepoContainer {
	logger := slog.NewLogger("whaaaaat?")
	assert.IsTrue(false, logger, "Something happened?")

	tagsRepo := initTagsRepo(connection)
	rolesRepo := project_repos.InitRolesRepo(connection)

	usersRepo := initUsersRepo(connection)

	projectsRepo := project_repos.InitProjectsRepo(connection)
	projectUsersRepo := project_repos.InitProjectUsersRepo(connection)
	projectTagsRepo := project_repos.InitProjectTagsRepo(connection)

	tasksRepo := initTasksRepo(connection)
	taskTagsRepo := initTaskTagsRepo(connection)

	return &RepoContainer{
		rolesRepo: rolesRepo,
		tagsRepo:  tagsRepo,

		usersRepo: usersRepo,

		projectsRepo:     projectsRepo,
		projectUsersRepo: projectUsersRepo,
		projectTagsRepo:  projectTagsRepo,

		tasksRepo:    tasksRepo,
		taskTagsRepo: taskTagsRepo,
	}
}

// NOTE: ENUMS
func (r *RepoContainer) GetTagsRepo() *TagsRepo {
	return r.tagsRepo
}

func (r *RepoContainer) GetRolesRepo() *project_repos.RolesRepo {
	return r.rolesRepo
}

// NOTE: Users
func (r *RepoContainer) GetUsersRepo() *UsersRepo {
	return r.usersRepo
}

// NOTE: Projects
func (r *RepoContainer) GetProjectsRepo() *project_repos.ProjectsRepo {
	return r.projectsRepo
}

func (r *RepoContainer) GetProjectUsersRepo() *project_repos.ProjectUsersRepo {
	return r.projectUsersRepo
}

func (r *RepoContainer) GetProjectTagsRepo() *project_repos.ProjectTagsRepo {
	return r.projectTagsRepo
}

// NOTE: Tasks
func (r *RepoContainer) GetTasksRepo() *TasksRepo {
	return r.tasksRepo
}

func (r *RepoContainer) GetTaskTagsRepo() *TasksRepo {
	return r.tasksRepo
}
