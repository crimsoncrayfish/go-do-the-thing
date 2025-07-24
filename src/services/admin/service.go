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

func (s *AdminService) ListUsers(currentUserId int64) ([]models.UserView, error) {
	s.logger.Debug("ListUsers called by userId: %d", currentUserId)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for ListUsers")
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to list users without admin rights", currentUserId)
		return nil, errors.New("only admins can view inactive users")
	}
	users, err := s.usersRepo.GetUsers()
	if err != nil {
		s.logger.Error(err, "failed to get inactive users")
		return nil, err
	}
	user_views := make([]models.UserView, len(users))
	for i, user := range users {
		user_views[i] = user.ToViewModel()
	}
	return user_views, nil
}

func (s *AdminService) GetUserById(currentUserId int64, user_id int64) (models.UserView, error) {
	s.logger.Debug("ListUsers called by userId: %d", currentUserId)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for ListUsers")
		return models.UserView{}, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to list users without admin rights", currentUserId)
		return models.UserView{}, errors.New("only admins can view inactive users")
	}

	user, err := s.usersRepo.GetUserById(user_id)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to get user by id - id: %d", user_id)
		return models.UserView{}, err
	}
	return user.ToViewModel(), nil
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
	return s.usersRepo.UpdateUserEnabled(userId, true)
}

func (s *AdminService) DeactivateUser(currentUserId, userId int64) (string, error) {
	s.logger.Debug("DeactivateUser called by userId: %d for userId: %d", currentUserId, userId)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for DeactivateUser")
		return "", err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to deactivate user %d without admin rights", currentUserId, userId)
		return "", errors.New("only admins can deactivate users")
	}
	return s.usersRepo.UpdateUserEnabled(userId, false)
}
