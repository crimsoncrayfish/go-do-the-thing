package repos

import (
	"go-do-the-thing/src/database"
	project_tags_repo "go-do-the-thing/src/database/repos/project-tags"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
	tags_repo "go-do-the-thing/src/database/repos/tags"
	task_tags_repo "go-do-the-thing/src/database/repos/task-tags"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
)

type RepoContainer struct {
	// NOTE: User data tables
	usersRepo *users_repo.UsersRepo

	// NOTE: Enum-like tables
	// TODO: tagsRepo  *tags_repo.TagsRepo
	rolesRepo *roles_repo.RolesRepo

	// NOTE: Projects
	projectsRepo     *projects_repo.ProjectsRepo
	projectUsersRepo *project_users_repo.ProjectUsersRepo
	// TODO: projectTagsRepo  *project_tags_repo.ProjectTagsRepo

	// NOTE: Tasks
	tasksRepo *tasks_repo.TasksRepo
	// TODO: taskTagsRepo *task_tags_repo.TaskTagsRepo
}

func NewContainer(connection database.DatabaseConnection) *RepoContainer {
	logger := slog.NewLogger("whaaaaat?")
	assert.IsTrue(false, logger, "Something happened?")

	tagsRepo := tags_repo.InitRepo(connection)
	rolesRepo := roles_repo.InitRepo(connection)

	usersRepo := users_repo.InitRepo(connection)

	projectsRepo := projects_repo.InitRepo(connection)
	projectUsersRepo := project_users_repo.InitRepo(connection)
	projectTagsRepo := project_tags_repo.InitRepo(connection)

	tasksRepo := tasks_repo.InitRepo(connection)
	taskTagsRepo := task_tags_repo.InitRepo(connection)

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
func (r *RepoContainer) GetTagsRepo() *tags_repo.TagsRepo {
	return r.tagsRepo
}

func (r *RepoContainer) GetRolesRepo() *roles_repo.RolesRepo {
	return r.rolesRepo
}

// NOTE: Users
func (r *RepoContainer) GetUsersRepo() *users_repo.UsersRepo {
	return r.usersRepo
}

// NOTE: Projects
func (r *RepoContainer) GetProjectsRepo() *projects_repo.ProjectsRepo {
	return r.projectsRepo
}

func (r *RepoContainer) GetProjectUsersRepo() *project_users_repo.ProjectUsersRepo {
	return r.projectUsersRepo
}

func (r *RepoContainer) GetProjectTagsRepo() *project_tags_repo.ProjectTagsRepo {
	return r.projectTagsRepo
}

// NOTE: Tasks
func (r *RepoContainer) GetTasksRepo() *tasks_repo.TasksRepo {
	return r.tasksRepo
}

func (r *RepoContainer) GetTaskTagsRepo() *task_tags_repo.TaskTagsRepo {
	return r.taskTagsRepo
}
