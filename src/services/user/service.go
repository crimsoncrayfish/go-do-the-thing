package user_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	usersRepo        users_repo.UsersRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
	logger           slog.Logger
}

const serviceSource = "UserService"

func SetupUserService(repo_container *repos.RepoContainer) UserService {
	return UserService{
		usersRepo:        *repo_container.GetUsersRepo(),
		projectUsersRepo: *repo_container.GetProjectUsersRepo(),
		logger:           slog.NewLogger(serviceSource),
	}
}

func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	s.logger.Debug("AuthenticateUser called - email: %s", email)
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to get user by email - email: %s", email)
		return nil, err
	}
	if !user.IsEnabled {
		s.logger.Warn("AuthenticateUser: user is not activated - email: %s", email)
		return nil, errors.New("login failed")
	}
	if user.IsDeleted {
		s.logger.Info("AuthenticateUser: user is deleted - email: %s", email)
		return nil, errors.New("login failed")
	}

	// Get password hash separately for security
	passwordHash, err := s.usersRepo.GetUserPassword(user.Id)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to get user password - email: %s, user_id: %d", email, user.Id)
		return nil, err
	}

	if !security.CheckPassword(password, passwordHash) {
		s.logger.Info("AuthenticateUser: invalid password - email: %s", email)
		return nil, errors.New("login failed")
	}
	// Generate new session ID
	sessionId := uuid.New().String()
	s.logger.Debug("AuthenticateUser: authentication succeeded - email: %s, user_id: %d", email, user.Id)
	// Update session in DB
	now := time.Now().UTC()
	err = s.usersRepo.UpdateSession(user.Id, sessionId, &now)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to update session - email: %s, user_id: %d", email, user.Id)
		return nil, err
	}
	user.SessionId = sessionId
	user.SessionStartTime = &now
	return user, nil
}

func (s *UserService) RegisterUser(name, email, password, password2 string) (*models.User, error) {
	if password != password2 {
		return nil, errors.New("passwords do not match")
	}
	user, _ := s.usersRepo.GetUserByEmail(email)
	if user != nil {
		return nil, errors.New("email already in use")
	}
	passwordHash, err := security.SetPassword(password)
	if err != nil {
		return nil, err
	}
	user = &models.User{
		Email:        email,
		FullName:     name,
		PasswordHash: passwordHash,
		IsDeleted:    false,
		IsAdmin:      false,
	}
	user_id, err := s.usersRepo.Create(user)
	if err != nil {
		return nil, err
	}
	user.Id = user_id
	return user, nil
}

func (s *UserService) LogoutUser(userId int64) error {
	return s.usersRepo.UpdateSession(userId, "", &time.Time{})
}

func (s *UserService) GetUserById(userId int64) (models.UserView, error) {
	user, err := s.usersRepo.GetUserById(userId)
	if err != nil {
		s.logger.Error(err, "Failed to get user with id %d", userId)
		return models.UserView{}, err
	}
	return user.ToViewModel(), nil
}

func (s *UserService) IsUserAdmin(userId int64) (bool, error) {
	user, err := s.usersRepo.GetUserById(userId)
	if err != nil {
		s.logger.Error(err, "Failed to get user with id %d", userId)
		return false, err
	}
	return user.IsAdmin, nil
}
