package admin_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"
)

type AdminService struct {
	usersRepo users_repo.UsersRepo
	logger    *logger.Logger
}

const serviceSource = "AdminService"

func SetupAdminService(repo_container *repos.RepoContainer) AdminService {
	log := logger.NewLogger(serviceSource)
	return AdminService{
		usersRepo: *repo_container.GetUsersRepo(),
		logger:    log,
	}
}

func (s *AdminService) ListUsers(currentUserId int64) ([]models.UserView, error) {
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error("failed to get current user for ListUsers", "error", err)
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User attempted to list users without admin rights", "currentUserId", currentUserId)
		return nil, errors.New("only admins can view inactive users")
	}
	users, err := s.usersRepo.GetUsers()
	if err != nil {
		s.logger.Error("failed to get inactive users", "error", err)
		return nil, err
	}
	user_views := make([]models.UserView, len(users))
	for i, user := range users {
		user_views[i] = user.ToViewModel()
	}
	return user_views, nil
}

func (s *AdminService) GetUserById(currentUserId int64, user_id int64) (models.UserView, error) {
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error("failed to get current user for GetUserById", "error", err)
		return models.UserView{}, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User attempted to get user without admin rights", "currentUserId", currentUserId)
		return models.UserView{}, errors.New("only admins can view inactive users")
	}

	user, err := s.usersRepo.GetUserById(user_id)
	if err != nil {
		s.logger.Error("failed to get user by id", "error", err, "user_id", user_id)
		return models.UserView{}, err
	}
	return user.ToViewModel(), nil
}

func (s *AdminService) ActivateUser(currentUserId, userId int64) (string, error) {
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error("failed to get current user for ActivateUser", "error", err)
		return "", err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User attempted to activate user without admin rights", "currentUserId", currentUserId, "user_id", userId)
		return "", errors.New("only admins can activate users")
	}
	return s.usersRepo.UpdateUserEnabled(userId, true)
}

func (s *AdminService) DeactivateUser(currentUserId, userId int64) (string, error) {
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error("failed to get current user for DeactivateUser", "error", err)
		return "", err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User attempted to deactivate user without admin rights", "currentUserId", currentUserId, "user_id", userId)
		return "", errors.New("only admins can deactivate users")
	}
	return s.usersRepo.UpdateUserEnabled(userId, false)
}
