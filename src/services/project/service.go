package projects_service

import (
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	roles_repo "go-do-the-thing/src/database/repos/roles"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type ProjectService struct {
	logger           slog.Logger
	projectRepo      projects_repo.ProjectsRepo
	usersRepo        users_repo.UsersRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
	rolesRepo        roles_repo.RolesRepo
}

const serviceSource = "ProjectService"

func SetupProjectService(repo_container *repos.RepoContainer) ProjectService {
	return ProjectService{
		logger:           slog.NewLogger(serviceSource),
		projectRepo:      *repo_container.GetProjectsRepo(),
		usersRepo:        *repo_container.GetUsersRepo(),
		projectUsersRepo: *repo_container.GetProjectUsersRepo(),
		rolesRepo:        *repo_container.GetRolesRepo(),
	}
}

func (s ProjectService) GetProjectView(id int64, currentUserId int64) (*models.ProjectView, error) {
	project, err := s.getProjectForUser(id, currentUserId)
	if err != nil {
		// NOTE: this should already be nicely formatted
		return nil, err
	}

	// NOTE: Success zone
	projectView, err := s.projectToViewModel(*project)
	if err != nil {
		// NOTE: this should already be nicely formatted
		return nil, err
	}
	return projectView, nil
}

func (s ProjectService) GetAllProjectsForUser(currentUserId int64) (projects []models.ProjectView, err error) {
	project_list, err := s.projectRepo.GetProjects(currentUserId)
	if err != nil {
		// NOTE: Should already be nicely formatted
		return nil, err
	}
	pl_v, err := s.projectListToViewModels(project_list)
	if err != nil {
		return nil, fmt.Errorf("failed to convert project list to project view list: %w", err)
	}
	return pl_v, nil
}

func (s ProjectService) DeleteProject(id, currentUserId int64) (hasProjects bool, err error) {
	err = s.projectRepo.DeleteProject(id, currentUserId)
	if err != nil {
		return false, err
	}

	// NOTE: Success zone
	projectCount, err := s.projectRepo.GetProjectCount(currentUserId)
	if err != nil {
		return false, err
	}
	return projectCount > 0, nil
}

func (s ProjectService) CreateProject(
	currentUserId, owner int64,
	name, description string,
	startDate, dueDate *database.SqLiteTime,
) (int64, error) {
	project := models.Project{
		Name:         name,
		Description:  description,
		Owner:        owner,
		StartDate:    startDate,
		DueDate:      dueDate,
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsComplete:   false,
		IsDeleted:    false,
	}

	id, err := s.projectRepo.Insert(currentUserId, project)
	if err != nil {
		return 0, err
	}

	inserted_id, err := s.projectUsersRepo.Insert(id, currentUserId, 0)
	if err != nil {
		return 0, err
	}
	return inserted_id, nil
}

func (s ProjectService) getProjectForUser(id int64, currentUserId int64) (*models.Project, error) {
	project, err := s.projectRepo.GetProject(id)
	if err != nil {
		// NOTE: this should already be nicely formatted
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) projectListToViewModels(projects []models.Project) (projectViews []models.ProjectView, err error) {
	projectViews = make([]models.ProjectView, len(projects))
	users := make(map[int64]*models.User)

	for i, project := range projects {
		var owner *models.User

		owner, ok := users[project.Owner]
		if !ok {
			owner, err = s.usersRepo.GetUserById(project.Owner)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[project.Owner] = owner
		}
		assert.NotNil(owner, serviceSource, fmt.Sprintf("project owner cant be nil - owner id %d", project.Owner))

		var createdBy *models.User
		createdBy, ok = users[project.CreatedBy]
		if !ok {
			createdBy, err = s.usersRepo.GetUserById(project.CreatedBy)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[project.CreatedBy] = createdBy
		}
		assert.NotNil(createdBy, serviceSource, fmt.Sprintf("project creator cant be nil - creator id %d", project.CreatedBy))

		var modifiedBy *models.User
		modifiedBy, ok = users[project.ModifiedBy]
		if !ok {
			modifiedBy, err = s.usersRepo.GetUserById(project.ModifiedBy)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[project.ModifiedBy] = modifiedBy
		}
		assert.NotNil(createdBy, serviceSource, fmt.Sprintf("project creator cant be nil - modifier id %d", project.ModifiedBy))

		// Convert to ViewModel
		projectViews[i] = project.ToViewModel(owner, createdBy, modifiedBy)
	}

	return projectViews, nil
}

func (s *ProjectService) projectToViewModel(project models.Project) (viewModel *models.ProjectView, err error) {
	users := make(map[int64]*models.User, 3)
	var owner *models.User
	owner, ok := users[project.Owner]
	s.logger.Debug("owner:%d", project.Owner)
	if !ok {
		owner, err = s.usersRepo.GetUserById(project.Owner)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[project.Owner] = owner
	}
	var created_by *models.User
	created_by, ok = users[project.CreatedBy]
	if !ok {
		created_by, err = s.usersRepo.GetUserById(project.CreatedBy)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[project.CreatedBy] = created_by
	}
	var modified_by *models.User
	modified_by, ok = users[project.ModifiedBy]
	if !ok {
		modified_by, err = s.usersRepo.GetUserById(project.ModifiedBy)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[project.ModifiedBy] = modified_by
	}
	s.logger.Debug("owner:%v", owner)
	projectView := project.ToViewModel(owner, created_by, modified_by)

	return &projectView, nil
}
