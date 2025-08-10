package admin_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	analytics_repo "go-do-the-thing/src/database/repos/analytics"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type AdminService struct {
	usersRepo     users_repo.UsersRepo
	analyticsRepo analytics_repo.AnalyticsRepo
	logger        slog.Logger
}

const serviceSource = "AdminService"

func SetupAdminService(repo_container *repos.RepoContainer) AdminService {
	return AdminService{
		usersRepo:     *repo_container.GetUsersRepo(),
		analyticsRepo: *repo_container.GetAnalyticsRepo(),
		logger:        slog.NewLogger(serviceSource),
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

func (s *AdminService) GetAnalyticsDashboard(currentUserId int64) (*models.AnalyticsDashboard, error) {
	s.logger.Debug("GetAnalyticsDashboard called by userId: %d", currentUserId)

	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for GetAnalyticsDashboard")
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to access analytics without admin rights", currentUserId)
		return nil, errors.New("only admins can view analytics")
	}

	userStats, err := s.analyticsRepo.GetUserStats()
	if err != nil {
		s.logger.Error(err, "failed to get user statistics")
		return nil, err
	}

	projectStats, err := s.analyticsRepo.GetProjectStats()
	if err != nil {
		s.logger.Error(err, "failed to get project statistics")
		return nil, err
	}

	taskStats, err := s.analyticsRepo.GetTaskStats()
	if err != nil {
		s.logger.Error(err, "failed to get task statistics")
		return nil, err
	}

	recentActivity, err := s.analyticsRepo.GetRecentActivity()
	if err != nil {
		s.logger.Error(err, "failed to get recent activity")
		return nil, err
	}

	dailyActivity, err := s.analyticsRepo.GetDailyActivity()
	if err != nil {
		s.logger.Error(err, "failed to get daily activity")
		return nil, err
	}

	dashboard := &models.AnalyticsDashboard{
		UserStats:      *userStats,
		ProjectStats:   *projectStats,
		TaskStats:      *taskStats,
		RecentActivity: *recentActivity,
		DailyActivity:  dailyActivity,
	}

	return dashboard, nil
}

func (s *AdminService) SearchUsers(currentUserId int64, searchTerm string) ([]models.UserView, error) {
	s.logger.Debug("SearchUsers called by userId: %d, searchTerm: %s", currentUserId, searchTerm)
	currentUser, err := s.usersRepo.GetUserById(currentUserId)
	if err != nil {
		s.logger.Error(err, "failed to get current user for SearchUsers")
		return nil, err
	}
	if currentUser == nil || !currentUser.IsAdmin {
		s.logger.Warn("User %d attempted to search users without admin rights", currentUserId)
		return nil, errors.New("only admins can search users")
	}

	if searchTerm == "" {
		return s.ListUsers(currentUserId)
	}

	users, err := s.usersRepo.SearchUsers(searchTerm)
	if err != nil {
		s.logger.Error(err, "failed to search users")
		return nil, err
	}
	user_views := make([]models.UserView, len(users))
	for i, user := range users {
		user_views[i] = user.ToViewModel()
	}
	return user_views, nil
}
