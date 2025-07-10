package admin_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type AdminService struct {
	usersRepo users_repo.UsersRepo
	logger    slog.Logger
}

const serviceSource = "AdminService"

func SetupAdminService(repo_container *repos.RepoContainer) AdminService {
	return AdminService{
		usersRepo: *repo_container.GetUsersRepo(),
		logger:    slog.NewLogger(serviceSource),
	}
}

func (s *AdminService) ListInactiveUsers(currentUserId int64) ([]models.UserView, error) {
	s.logger.Debug("ListInactiveUsers called by userId: %d", currentUserId)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for ListInactiveUsers")
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to list inactive users without admin rights", currentUserId)
		return nil, errors.New("only admins can view inactive users")
	}
	users, err := s.usersRepo.GetInactiveUsers()
	if err != nil {
		s.logger.Error(err, "failed to get inactive users")
		return nil, err
	}
	inactive := make([]models.UserView, 0, len(users))
	for _, user := range users {
		inactive = append(inactive, user.ToViewModel())
	}
	return inactive, nil
}

func (s *AdminService) ActivateUser(currentUserId, userId int64) (string, error) {
	s.logger.Debug("ActivateUser called by userId: %d for userId: %d", currentUserId, userId)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for ActivateUser")
		return "", err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to activate user %d without admin rights", currentUserId, userId)
		return "", errors.New("only admins can activate users")
	}
	return s.usersRepo.ActivateUser(userId)
}
