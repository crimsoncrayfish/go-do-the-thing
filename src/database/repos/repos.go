package repos

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/slog"
)

type RepoContainer struct {
	usersRepo *UsersRepo
	tasksRepo *TasksRepo
}

func NewContainer(connection database.DatabaseConnection) (*RepoContainer, error) {
	logger := slog.NewLogger("Repository Setup")
	usersRepo, err := InitUsersRepo(connection)
	if err != nil {
		logger.Error(err, "could not initialise the users repository")
		return nil, err
	}
	tasksRepo, err := InitTasksRepo(connection)
	if err != nil {
		logger.Error(err, "could not initialise the tasks repository")
		return nil, err
	}
	return &RepoContainer{usersRepo: usersRepo, tasksRepo: tasksRepo}, nil
}

func (r *RepoContainer) GetUsersRepo() *UsersRepo {
	return r.usersRepo
}

func (r *RepoContainer) GetTasksRepo() *TasksRepo {
	return r.tasksRepo
}
