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

func (s *UserService) AuthenticateUser(email, password string) (*models.User, string, error) {
	s.logger.Info("AuthenticateUser called - email: %s", email)
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to get user by email - email: %s", email)
		return nil, "", err
	}
	if user.IsDeleted {
		s.logger.Info("AuthenticateUser: user is deleted - email: %s", email)
		return nil, "", errors.New("user is deleted")
	}

	// Get password hash separately for security
	passwordHash, err := s.usersRepo.GetUserPassword(user.Id)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to get user password - email: %s, user_id: %d", email, user.Id)
		return nil, "", err
	}

	if !security.CheckPassword(password, passwordHash) {
		s.logger.Info("AuthenticateUser: invalid password - email: %s", email)
		return nil, "", errors.New("invalid password")
	}
	// Generate new session ID
	sessionId := uuid.New().String()
	s.logger.Info("AuthenticateUser: authentication succeeded - email: %s, user_id: %d", email, user.Id)
	// Update session in DB
	now := time.Now().UTC()
	err = s.usersRepo.UpdateSession(user.Id, sessionId, &now)
	if err != nil {
		s.logger.Error(err, "AuthenticateUser: failed to update session - email: %s, user_id: %d", email, user.Id)
		return nil, "", err
	}
	return user, sessionId, nil
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
