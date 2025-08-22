// Package taskservice implements the service layer for interacting with tasks
package taskservice

import (
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	tasktagsrepo "go-do-the-thing/src/database/repos/task-tags"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"
	"sort"
	"time"
)

type TaskService struct {
	tasksRepo        tasks_repo.TasksRepo
	taskTagsRepo     tasktagsrepo.TaskTagsRepo
	usersRepo        users_repo.UsersRepo
	projectRepo      projects_repo.ProjectsRepo
	projectUsersRepo project_users_repo.ProjectUsersRepo
	logger           *logger.Logger
}

const serviceSource = "TaskService"

func SetupTaskService(repoContainer *repos.RepoContainer) TaskService {
	log := logger.NewLogger(serviceSource)
	return TaskService{
		tasksRepo:        *repoContainer.GetTasksRepo(),
		taskTagsRepo:     *repoContainer.GetTaskTagsRepo(),
		usersRepo:        *repoContainer.GetUsersRepo(),
		projectRepo:      *repoContainer.GetProjectsRepo(),
		projectUsersRepo: *repoContainer.GetProjectUsersRepo(),
		logger:           log,
	}
}

func (s *TaskService) CreateTask(userID, projectID int64, name, description string, dueDate *time.Time, taskTags []int64) (int64, error) {
	err := s.UserBelongsToProject(userID, projectID)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      dueDate,
		AssignedTo:   userID, // TODO: need to update this
		CreatedBy:    userID,
		CreatedDate:  &now,
		ModifiedBy:   userID,
		ModifiedDate: &now,
		Project:      projectID,
		IsDeleted:    false,
	}

	id, err := s.tasksRepo.InsertTask(task)
	if err != nil {
		return 0, err
	}
	s.logger.Info("task created", "taskID", id, "name", name, "projectID", projectID, "userID", userID)

	if err = s.taskTagsRepo.SetTaskTags(projectID, id, taskTags); err != nil {
		return id, err
	}

	return id, nil
}

func (s *TaskService) UpdateTask(userID, taskID, projectID int64, name, description string, dueDate *time.Time, assignedTo int64, taskTagIDs []int64) error {
	err := s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return err
	}
	err = s.UserBelongsToProject(userID, projectID)
	if err != nil {
		return err
	}
	now := time.Now()
	task := models.Task{
		ID:           taskID,
		Name:         name,
		Description:  description,
		DueDate:      dueDate,
		AssignedTo:   assignedTo, // TODO: need to update this
		ModifiedBy:   userID,
		ModifiedDate: &now,
		Project:      projectID,
		IsDeleted:    false,
	}
	err = s.tasksRepo.UpdateTask(task)
	if err != nil {
		return err
	}

	if err = s.taskTagsRepo.SetTaskTags(projectID, taskID, taskTagIDs); err != nil {
		return err
	}

	return nil
}

func (s *TaskService) UpdateTaskStatus(userID, taskID int64) error {
	err := s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return err
	}

	task, err := s.tasksRepo.GetTask(taskID)
	if err != nil {
		return err
	}

	if task.AssignedTo != userID {
		return errors.NewAccessDenied("task", "not user's task")
	}

	task.ToggleStatus(userID)

	return s.tasksRepo.UpdateTaskStatus(taskID, task.CompleteDate, int64(task.Status), userID)
}

func (s *TaskService) DeleteTask(userID, taskID int64) error {
	err := s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return err
	}

	err = s.tasksRepo.DeleteTask(taskID, userID)
	if err != nil {
		return err
	}
	s.logger.Info("task deleted", "taskID", taskID, "userID", userID)
	return nil
}

func (s *TaskService) RestoreTask(userID, taskID int64) error {
	err := s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return err
	}

	return s.tasksRepo.RestoreTask(taskID, userID)
}

func (s *TaskService) GetTaskView(id, userID int64) (*models.TaskView, error) {
	task, err := s.tasksRepo.GetTask(id)
	if err != nil {
		return nil, err
	}
	err = s.UserBelongsToProject(userID, task.Project)
	if err != nil {
		return nil, err
	}
	taskTags, err := s.taskTagsRepo.GetTagsByTaskID(id)
	if err != nil {
		return nil, err
	}
	task.Tags = taskTags

	return s.taskToViewModel(task)
}

func (s *TaskService) Search(userID int64, searchText string, projectID int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.Search(userID, searchText, projectID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		taskTags, err := s.taskTagsRepo.GetTagsByTaskID(task.ID)
		if err != nil {
			return nil, err
		}
		task.Tags = taskTags
	}
	return s.taskListToViewModels(tasks)
}

func (s *TaskService) GetTaskViewList(userID int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.GetTasksForUser(userID)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		taskTags, err := s.taskTagsRepo.GetTagsByTaskID(task.ID)
		if err != nil {
			return nil, err
		}
		task.Tags = taskTags
	}
	return s.taskListToViewModels(tasks)
}

func (s *TaskService) GetProjectTaskViewList(userID, projectID int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.GetTasksForUserAndProject(userID, projectID)
	if err != nil {
		return nil, err
	}
	return s.taskListToViewModels(tasks)
}

func (s *TaskService) GetTaskCount(userID int64) (int64, error) {
	taskCount, err := s.tasksRepo.GetTasksCount(userID)
	if err != nil {
		return 0, err
	}
	return taskCount, nil
}

func (s *TaskService) taskListToViewModels(tasks []*models.Task) (taskViews []*models.TaskView, err error) {
	taskViews = make([]*models.TaskView, len(tasks))
	users := make(map[int64]*models.User)
	projects := make(map[int64]*models.Project)
	for i, task := range tasks {
		var assignedTo *models.User

		assignedTo, ok := users[task.AssignedTo]
		if !ok {
			assignedTo, err = s.usersRepo.GetUserById(task.AssignedTo)
			if err != nil {
				return nil, err
			}
			users[task.AssignedTo] = assignedTo
		}

		var createdBy *models.User
		createdBy, ok = users[task.CreatedBy]
		if !ok {
			createdBy, err = s.usersRepo.GetUserById(task.CreatedBy)
			if err != nil {
				return nil, err
			}
			users[task.ModifiedBy] = createdBy
		}

		var modifiedBy *models.User
		modifiedBy, ok = users[task.ModifiedBy]
		if !ok {
			modifiedBy, err = s.usersRepo.GetUserById(task.ModifiedBy)
			if err != nil {
				return nil, err
			}
			users[task.ModifiedBy] = modifiedBy
		}

		var project *models.Project
		project, ok = projects[task.Project]
		if !ok {
			project, err = s.projectRepo.GetProject(task.Project)
			if err != nil {
				return nil, err
			}
			projects[task.Project] = project
		}

		view, err := task.ToViewModel(assignedTo, createdBy, modifiedBy, *project)
		if err != nil {
			return nil, err
		}
		taskViews[i] = view
	}

	sort.Slice(taskViews, func(i, j int) bool {
		if taskViews[i].Status != taskViews[j].Status {
			return taskViews[i].Status < taskViews[j].Status
		}
		return taskViews[i].DueDate.Before(*taskViews[j].DueDate)
	})
	return taskViews, nil
}

func (s *TaskService) taskToViewModel(task *models.Task) (view *models.TaskView, err error) {
	users := make(map[int64]*models.User, 3)
	var assignedTo *models.User
	assignedTo, ok := users[task.AssignedTo]
	if !ok {
		assignedTo, err = s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			return nil, err
		}
		users[task.AssignedTo] = assignedTo
	}

	var createdBy *models.User
	createdBy, ok = users[task.CreatedBy]
	if !ok {
		createdBy, err = s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			return nil, err
		}
		users[task.AssignedTo] = createdBy
	}

	var modifiedBy *models.User
	modifiedBy, ok = users[task.ModifiedBy]
	if !ok {
		modifiedBy, err = s.usersRepo.GetUserById(task.ModifiedBy)
		if err != nil {
			return nil, err
		}
		users[task.ModifiedBy] = modifiedBy
	}

	project, err := s.projectRepo.GetProject(task.Project)
	if err != nil {
		return nil, err
	}
	view, err = task.ToViewModel(assignedTo, createdBy, modifiedBy, *project)
	if err != nil {
		return nil, err
	}

	return view, nil
}

func (s *TaskService) userBelongsToTaskProject(userID, taskID int64) (err error) {
	task, err := s.tasksRepo.GetTask(taskID)
	if err != nil {
		return err
	}
	return s.UserBelongsToProject(userID, task.Project)
}

func (s *TaskService) UserBelongsToProject(userID, projectID int64) (err error) {
	roles, err := s.projectUsersRepo.GetProjectUserRoles(projectID, userID)
	if err != nil {
		return err
	}
	if len(roles) == 0 {
		return errors.NewAccessDenied("project", "user %d does not belong to project %d", userID, projectID)
	}
	return nil
}
