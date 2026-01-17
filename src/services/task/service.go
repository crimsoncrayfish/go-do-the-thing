// Package taskservice implements the service layer for interacting with tasks
package taskservice

import (
	"go-do-the-thing/src/database/repos"
	project_users_repo "go-do-the-thing/src/database/repos/project-users"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	tasktagsrepo "go-do-the-thing/src/database/repos/task-tags"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	timeeventsrepo "go-do-the-thing/src/database/repos/time-events"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"
	fm "go-do-the-thing/src/models/forms"
	"sort"
	"time"
)

type TaskService struct {
	tasksRepo        tasks_repo.TasksRepo
	taskTagsRepo     tasktagsrepo.TaskTagsRepo
	taskEventsRepo   timeeventsrepo.TaskEventsRepo
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
		taskEventsRepo:   *repoContainer.GetTaskEventsRepo(),
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

func (s *TaskService) Search(userID int64, searchText string, projectID int64, status fm.ItemStatusFilter, tagIDs []int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.Search(userID, searchText, projectID, status.String(), tagIDs)
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

		// Populate time tracking info
		inProgress, timeSpent, err := s.GetTaskTimeInfo(task.ID)
		if err != nil {
			s.logger.Warn("failed to get time info for task", "taskID", task.ID, "error", err)
			// Don't fail the whole request, just leave time info empty
		} else {
			view.InProgress = inProgress
			view.TimeSpent = time.Duration(timeSpent) * time.Second
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

	// Populate time tracking info
	inProgress, timeSpent, err := s.GetTaskTimeInfo(task.ID)
	if err != nil {
		s.logger.Warn("failed to get time info for task", "taskID", task.ID, "error", err)
		// Don't fail the whole request, just leave time info empty
	} else {
		view.InProgress = inProgress
		view.TimeSpent = time.Duration(timeSpent) * time.Second
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

// StartTimeTracking begins time tracking for a task.
// If another task is currently in progress for this user, it will be stopped first.
// Returns the ID of the previously tracked task (0 if none).
func (s *TaskService) StartTimeTracking(userID, taskID int64) (previousTaskID int64, err error) {
	err = s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return 0, err
	}

	// Check if this specific task is already in progress
	inProgress, err := s.taskEventsRepo.IsTaskInProgress(taskID)
	if err != nil {
		return 0, err
	}
	if inProgress {
		// Already tracking this task, nothing to do
		return 0, nil
	}

	// Check if user has any other task in progress and stop it
	previousTaskID, err = s.stopAnyInProgressTask(userID)
	if err != nil {
		return 0, err
	}

	// Start tracking the new task
	_, err = s.taskEventsRepo.InsertTimeEvent(taskID, userID, models.EventTypeTimeStart)
	if err != nil {
		return previousTaskID, err
	}

	s.logger.Info("time tracking started", "taskID", taskID, "userID", userID, "previousTaskID", previousTaskID)
	return previousTaskID, nil
}

// stopAnyInProgressTask finds and stops any task currently in progress for the user.
// Returns the task ID that was stopped (0 if none).
func (s *TaskService) stopAnyInProgressTask(userID int64) (int64, error) {
	// Get all tasks for the user and check which one is in progress
	tasks, err := s.tasksRepo.GetTasksForUser(userID)
	if err != nil {
		return 0, err
	}

	for _, task := range tasks {
		inProgress, err := s.taskEventsRepo.IsTaskInProgress(task.ID)
		if err != nil {
			continue // Skip on error, try next task
		}
		if inProgress {
			// Stop this task
			_, err = s.taskEventsRepo.InsertTimeEvent(task.ID, userID, models.EventTypeTimeStop)
			if err != nil {
				return 0, err
			}
			s.logger.Info("auto-stopped previous task", "taskID", task.ID, "userID", userID)
			return task.ID, nil
		}
	}

	return 0, nil
}

// StopTimeTracking stops time tracking for a task
func (s *TaskService) StopTimeTracking(userID, taskID int64) error {
	err := s.userBelongsToTaskProject(userID, taskID)
	if err != nil {
		return err
	}

	// Check if task is actually in progress
	inProgress, err := s.taskEventsRepo.IsTaskInProgress(taskID)
	if err != nil {
		return err
	}
	if !inProgress {
		return errors.New(errors.ErrValidationFailed, "task is not currently being tracked")
	}

	_, err = s.taskEventsRepo.InsertTimeEvent(taskID, userID, models.EventTypeTimeStop)
	if err != nil {
		return err
	}

	s.logger.Info("time tracking stopped", "taskID", taskID, "userID", userID)
	return nil
}

// GetTaskTimeInfo returns time tracking info for a task (in progress status and time spent)
func (s *TaskService) GetTaskTimeInfo(taskID int64) (inProgress bool, timeSpentSeconds int64, err error) {
	inProgress, err = s.taskEventsRepo.IsTaskInProgress(taskID)
	if err != nil {
		return false, 0, err
	}

	timeSpentSeconds, err = s.taskEventsRepo.GetTaskTimeSpent(taskID)
	if err != nil {
		return inProgress, 0, err
	}

	return inProgress, timeSpentSeconds, nil
}
