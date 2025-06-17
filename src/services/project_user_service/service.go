package project_user_service

import (
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
)

type ProjectUserService struct {
	logger           slog.Logger
	projectUsersRepo project_users_repo.ProjectUsersRepo
}

const serviceSource = "ProjectService"

func SetupProjectUserService(repo_container *repos.RepoContainer) ProjectUserService {
	return ProjectUserService{
		logger:           slog.NewLogger(serviceSource),
		projectUsersRepo: *repo_container.GetProjectUsersRepo(),
	}
}

func (s *ProjectUserService) UserBelongsToProject(user_id, project_id int64) (err error) {
	roles, err := s.projectUsersRepo.GetProjectUserRoles(project_id, user_id)
	if err != nil {
		// NOTE: Errors from repo are wrapped
		return err
	}
	if len(roles) == 0 {
		return errors.New(errors.ErrAccessDenied, "permission denied: user %d does not belong to project %d", user_id, project_id)
	}
	return nil
}
