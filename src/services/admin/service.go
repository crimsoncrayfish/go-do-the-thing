package admin_service

import (
	"go-do-the-thing/src/database/repos"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/errors"
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

func (s *AdminService) requireAdmin(currentUserId int64) (*models.User, error) {
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error("admin check failed", "error", err, "user_id", currentUserId)
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("non-admin attempted admin action", "user_id", currentUserId)
		return nil, errors.NewAccessDenied("admin functions", "admin role required")
	}
	return currentUser, nil
}

func (s *AdminService) ListUsers(currentUserId int64) ([]models.UserView, error) {
	_, err := s.requireAdmin(currentUserId)
	if err != nil {
		return nil, err
	}
	users, err := s.usersRepo.GetUsers()
	if err != nil {
		return nil, err
	}
	user_views := make([]models.UserView, len(users))
	for i, user := range users {
		user_views[i] = user.ToViewModel()
	}
	return user_views, nil
}

func (s *AdminService) GetUserById(currentUserId int64, user_id int64) (models.UserView, error) {
	_, err := s.requireAdmin(currentUserId)
	if err != nil {
		return models.UserView{}, err
	}

	user, err := s.usersRepo.GetUserById(user_id)
	if err != nil {
		return models.UserView{}, err
	}
	return user.ToViewModel(), nil
}

func (s *AdminService) ActivateUser(currentUserId, userId int64) (string, error) {
	_, err := s.requireAdmin(currentUserId)
	if err != nil {
		return "", err
	}
	s.logger.Info("admin activating user", "admin_id", currentUserId, "target_user_id", userId)
	return s.usersRepo.UpdateUserEnabled(userId, true)
}

func (s *AdminService) DeactivateUser(currentUserId, userId int64) (string, error) {
	_, err := s.requireAdmin(currentUserId)
	if err != nil {
		return "", err
	}
	s.logger.Info("admin deactivating user", "admin_id", currentUserId, "target_user_id", userId)
	return s.usersRepo.UpdateUserEnabled(userId, false)
}
