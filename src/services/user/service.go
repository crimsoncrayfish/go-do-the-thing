package user_service

import (
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	users_repo "go-do-the-thing/src/database/repos/users"
)

type UserService struct {
	usersRepo        users_repo.UsersRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
}

const serviceSource = "UserService"

func SetupUserService(repo_container *repos.RepoContainer) *UserService {
	return &UserService{
		usersRepo:        *repo_container.GetUsersRepo(),
		projectUsersRepo: *repo_container.GetProjectUsersRepo(),
	}
}
