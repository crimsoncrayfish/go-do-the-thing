package repos

import (
	"go-do-the-thing/src/database"
	project_tags_repo "go-do-the-thing/src/database/repos/project-tags"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
	tasktagsrepo "go-do-the-thing/src/database/repos/task-tags"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	timeeventsrepo "go-do-the-thing/src/database/repos/time-events"
	users_repo "go-do-the-thing/src/database/repos/users"
)

type RepoContainer struct {
	// NOTE: User data tables
	usersRepo *users_repo.UsersRepo

	rolesRepo *roles_repo.RolesRepo

	// NOTE: Projects
	projectsRepo     *projects_repo.ProjectsRepo
	projectUsersRepo *project_users_repo.ProjectUsersRepo
	projectTagsRepo  *project_tags_repo.ProjectTagsRepo

	// NOTE: Tasks
	tasksRepo      *tasks_repo.TasksRepo
	taskTagsRepo   *tasktagsrepo.TaskTagsRepo
	taskEventsRepo *timeeventsrepo.TaskEventsRepo
}

func NewContainer(connection *database.DatabaseConnection) *RepoContainer {
	rolesRepo := roles_repo.InitRepo(connection)

	usersRepo := users_repo.InitRepo(connection)

	projectsRepo := projects_repo.InitRepo(connection)
	projectUsersRepo := project_users_repo.InitRepo(connection)
	projectTagsRepo := project_tags_repo.InitRepo(connection)

	tasksRepo := tasks_repo.InitRepo(connection)
	taskTagsRepo := tasktagsrepo.InitRepo(connection)
	taskEventsRepo := timeeventsrepo.InitRepo(connection)

	return &RepoContainer{
		rolesRepo: rolesRepo,

		usersRepo: usersRepo,

		projectsRepo:     projectsRepo,
		projectUsersRepo: projectUsersRepo,
		projectTagsRepo:  projectTagsRepo,

		tasksRepo:      tasksRepo,
		taskTagsRepo:   taskTagsRepo,
		taskEventsRepo: taskEventsRepo,
	}
}

func (r *RepoContainer) GetRolesRepo() *roles_repo.RolesRepo {
	return r.rolesRepo
}

func (r *RepoContainer) GetUsersRepo() *users_repo.UsersRepo {
	return r.usersRepo
}

func (r *RepoContainer) GetProjectsRepo() *projects_repo.ProjectsRepo {
	return r.projectsRepo
}

func (r *RepoContainer) GetProjectUsersRepo() *project_users_repo.ProjectUsersRepo {
	return r.projectUsersRepo
}

func (r *RepoContainer) GetProjectTagsRepo() *project_tags_repo.ProjectTagsRepo {
	return r.projectTagsRepo
}

func (r *RepoContainer) GetTasksRepo() *tasks_repo.TasksRepo {
	return r.tasksRepo
}

func (r *RepoContainer) GetTaskTagsRepo() *tasktagsrepo.TaskTagsRepo {
	return r.taskTagsRepo
}

func (r *RepoContainer) GetTaskEventsRepo() *timeeventsrepo.TaskEventsRepo {
	return r.taskEventsRepo
}
