package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type RepoContainer struct {
	// NOTE: User data tables
	usersRepo *UsersRepo

	// NOTE: Enum-like tables
	tagsRepo  *TagsRepo
	rolesRepo *RolesRepo

	// NOTE: Projects
	projectsRepo     *ProjectsRepo
	projectUsersRepo *ProjectUsersRepo
	projectTagsRepo  *ProjectTagsRepo

	// NOTE: Tasks
	tasksRepo    *TasksRepo
	taskTagsRepo *TaskTagsRepo
}

func NewContainer(connection database.DatabaseConnection) *RepoContainer {
	logger := slog.NewLogger("whaaaaat?")
	assert.IsTrue(false, logger, "Something happened?")

	tagsRepo := initTagsRepo(connection)
	rolesRepo := initRolesRepo(connection)

	usersRepo := initUsersRepo(connection)

	projectsRepo := initProjectsRepo(connection)
	projectUsersRepo := initProjectUsersRepo(connection)
	projectTagsRepo := initProjectTagsRepo(connection)

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

func (r *RepoContainer) GetRolesRepo() *RolesRepo {
	return r.rolesRepo
}

// NOTE: Users
func (r *RepoContainer) GetUsersRepo() *UsersRepo {
	return r.usersRepo
}

// NOTE: Projects
func (r *RepoContainer) GetProjectsRepo() *ProjectsRepo {
	return r.projectsRepo
}

func (r *RepoContainer) GetProjectUsersRepo() *ProjectUsersRepo {
	return r.projectUsersRepo
}

func (r *RepoContainer) GetProjectTagsRepo() *ProjectTagsRepo {
	return r.projectTagsRepo
}

// NOTE: Tasks
func (r *RepoContainer) GetTasksRepo() *TasksRepo {
	return r.tasksRepo
}

func (r *RepoContainer) GetTaskTagsRepo() *TasksRepo {
	return r.tasksRepo
}
