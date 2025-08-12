package user_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/models"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type UserService struct {
	usersRepo        users_repo.UsersRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
	logger           *slog.Logger
}

const serviceSource = "UserService"

func SetupUserService(repo_container *repos.RepoContainer) UserService {
	logger := slog.With("source", serviceSource)
	return UserService{
		usersRepo:        *repo_container.GetUsersRepo(),
		projectUsersRepo: *repo_container.GetProjectUsersRepo(),
		logger:           logger,
	}
}

func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		s.logger.Error("AuthenticateUser: failed to get user by email", "error", err, "email", email)
		return nil, err
	}
	if user.IsDeleted {
		s.logger.Info("AuthenticateUser: user is deleted", "email", email)
		return nil, errors.New("login failed")
	}

	passwordHash, err := s.usersRepo.GetUserPassword(user.Id)
	if err != nil {
		s.logger.Error("AuthenticateUser: failed to get user password", "error", err, "email", email, "user_id", user.Id)
		return nil, err
	}

	if !security.CheckPassword(password, passwordHash) {
		s.logger.Info("AuthenticateUser: invalid password", "email", email)
		return nil, errors.New("login failed")
	}
	if !user.IsEnabled {
		s.logger.Warn("AuthenticateUser: user is not activated", "email", email)
		return nil, errors.New("user not enabled")
	}
	sessionId := uuid.New().String()
	now := time.Now().UTC()
	err = s.usersRepo.UpdateSession(user.Id, sessionId, &now)
	if err != nil {
		s.logger.Error("AuthenticateUser: failed to update session", "error", err, "email", email, "user_id", user.Id)
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
		s.logger.Error("Failed to get user", "error", err, "userId", userId)
		return models.UserView{}, err
	}
	return user.ToViewModel(), nil
}

func (s *UserService) IsUserAdmin(userId int64) (bool, error) {
	user, err := s.usersRepo.GetUserById(userId)
	if err != nil {
		s.logger.Error("Failed to get user", "error", err, "userId", userId)
		return false, err
	}
	return user.IsAdmin, nil
}
