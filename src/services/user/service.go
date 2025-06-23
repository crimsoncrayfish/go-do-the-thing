package user_service

import (
	"errors"
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/security"
	"go-do-the-thing/src/models"
	"time"
	"github.com/google/uuid"
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

func (s *UserService) AuthenticateUser(email, password string) (*models.User, string, error) {
	user, err := s.usersRepo.GetUserByEmail(email)
	if err != nil {
		return nil, "", err
	}
	passwordHash, err := s.usersRepo.GetUserPassword(user.Id)
	if err != nil {
		return nil, "", err
	}
	if !security.CheckPassword(password, passwordHash) {
		return nil, "", errors.New("invalid password")
	}
	now := time.Now()
	user.SessionId = uuid.New().String()
	user.SessionStartTime = &now
	if err := s.usersRepo.UpdateSession(user.Id, user.SessionId, user.SessionStartTime); err != nil {
		return nil, "", err
	}
	return user, user.SessionId, nil
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
	_, err = s.usersRepo.Create(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) LogoutUser(userId int64) error {
	return s.usersRepo.UpdateSession(userId, "", &time.Time{})
}
