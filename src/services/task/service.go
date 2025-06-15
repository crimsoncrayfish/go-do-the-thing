package task_service

import (
	"fmt"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/database/repos"
	projects_repo "go-do-the-thing/src/database/repos/projects"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/assert"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	project_user_service "go-do-the-thing/src/services/project_user_service"
	"sort"
	"time"
)

type TaskService struct {
	tasksRepo           tasks_repo.TasksRepo
	usersRepo           users_repo.UsersRepo
	projectRepo         projects_repo.ProjectsRepo
	projectUsersService project_user_service.ProjectUserService
}

const serviceSource = "TaskService"

func SetupTaskService(repo_container *repos.RepoContainer) TaskService {
	return TaskService{
		tasksRepo:           *repo_container.GetTasksRepo(),
		usersRepo:           *repo_container.GetUsersRepo(),
		projectRepo:         *repo_container.GetProjectsRepo(),
		projectUsersService: project_user_service.SetupProjectUserService(repo_container),
	}
}

func (s *TaskService) CreateTask(user_id, project_id int64, name, description string, due_date *time.Time) (int64, error) {
	// NOTE: Does this user belong to the current project
	err := s.projectUsersService.UserBelongsToProject(user_id, project_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return 0, err
	}
	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      due_date,
		AssignedTo:   user_id, // TODO: need to update this
		CreatedBy:    user_id,
		CreatedDate:  time.Now(),
		ModifiedBy:   user_id,
		ModifiedDate: time.Now(),
		Project:      project_id,
		IsDeleted:    false,
	}

	id, err := s.tasksRepo.InsertItem(task)
	if err != nil {
		// NOTE: Errors from repo are wrapped
		return 0, err
	}
	return id, nil
}

func (s *TaskService) UpdateTask(user_id, task_id, project_id int64, name, description string, due_date *time.Time, assigned_to int64) error {
	// NOTE: Does this user belong to the current project
	err := s.userBelongsToTaskProject(user_id, task_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return err
	}
	err = s.projectUsersService.UserBelongsToProject(user_id, project_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return err
	}
	task := models.Task{
		Id:           task_id,
		Name:         name,
		Description:  description,
		DueDate:      due_date,
		AssignedTo:   assigned_to, // TODO: need to update this
		ModifiedBy:   user_id,
		ModifiedDate: time.Now(),
		Project:      project_id,
		IsDeleted:    false,
	}
	err = s.tasksRepo.UpdateItem(task)
	if err != nil {
		return err
	}

	return nil
}

func (s *TaskService) UpdateTaskStatus(user_id, task_id int64) error {
	err := s.userBelongsToTaskProject(user_id, task_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return err
	}

	task, err := s.tasksRepo.GetItem(task_id)
	if err != nil {
		// NOTE: Take action
		return err
	}

	if task.AssignedTo != user_id {
		return errors.New(errors.ErrAccessDenied, "not user's task")
	}

	// NOTE: Take action
	task.ToggleStatus(user_id)

	return s.tasksRepo.UpdateItemStatus(task_id, task.CompleteDate, int64(task.Status), user_id)
}

func (s *TaskService) DeleteTask(user_id, task_id int64) error {
	// NOTE: Does this user belong to the current project
	err := s.userBelongsToTaskProject(user_id, task_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return err
	}

	return s.tasksRepo.DeleteItem(task_id, user_id)
}

func (s *TaskService) RestoreTask(user_id, task_id int64) error {
	// NOTE: Does this user belong to the current project
	err := s.userBelongsToTaskProject(user_id, task_id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return err
	}

	return s.tasksRepo.RestoreItem(task_id, user_id)
}

func (s *TaskService) GetTaskView(id, user_id int64) (*models.TaskView, error) {
	task, err := s.tasksRepo.GetItem(id)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return nil, err
	}
	err = s.projectUsersService.UserBelongsToProject(user_id, task.Project)
	if err != nil {
		// NOTE: Errors from function already wrapped
		return nil, err
	}

	return s.taskToViewModel(task)
}

func (s *TaskService) GetTaskViewList(user_id int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.GetItemsForUser(user_id)
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Status != tasks[j].Status {
			return tasks[i].Status < tasks[j].Status
		}
		return tasks[i].DueDate.Before(tasks[j].DueDate)
	})
	return s.taskListToViewModels(tasks)
}

func (s *TaskService) GetProjectTaskViewList(user_id, project_id int64) ([]*models.TaskView, error) {
	tasks, err := s.tasksRepo.GetItemsForUserAndProject(user_id, project_id)
	if err != nil {
		return nil, err
	}
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Status != tasks[j].Status {
			return tasks[i].Status < tasks[j].Status
		}
		return tasks[i].DueDate.Before(tasks[j].DueDate)
	})
	return s.taskListToViewModels(tasks)
}

func (s *TaskService) GetTaskCount(user_id int64) (int64, error) {
	task_count, err := s.tasksRepo.GetItemsCount(user_id)
	if err != nil {
		return 0, err
	}

	return task_count, nil
}

func (s *TaskService) taskListToViewModels(tasks []*models.Task) (taskViews []*models.TaskView, err error) {
	taskViews = make([]*models.TaskView, len(tasks))
	users := make(map[int64]*models.User)
	projects := make(map[int64]*models.Project)

	for i, task := range tasks {
		var assigned_to *models.User

		assigned_to, ok := users[task.AssignedTo]
		if !ok {
			assigned_to, err = s.usersRepo.GetUserById(task.AssignedTo)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[task.AssignedTo] = assigned_to
		}
		assert.NotNil(assigned_to, serviceSource, fmt.Sprintf("task assignee cant be nil - assigned to id %d", task.AssignedTo))

		var created_by *models.User

		created_by, ok = users[task.CreatedBy]
		if !ok {
			created_by, err = s.usersRepo.GetUserById(task.CreatedBy)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[task.ModifiedBy] = created_by
		}
		assert.NotNil(created_by, serviceSource, fmt.Sprintf("task created by user cant be nil - created by id %d", task.CreatedBy))

		var modified_by *models.User

		modified_by, ok = users[task.ModifiedBy]
		if !ok {
			modified_by, err = s.usersRepo.GetUserById(task.ModifiedBy)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			users[task.ModifiedBy] = modified_by
		}
		assert.NotNil(modified_by, serviceSource, fmt.Sprintf("task modified by user cant be nil - modified by id %d", task.ModifiedBy))

		var project *models.Project
		project, ok = projects[task.Project]
		if !ok {
			project, err = s.projectRepo.GetProject(task.Project)
			if err != nil {
				// NOTE: this should already be nicely formatted
				return nil, err
			}
			projects[task.Project] = project

		}

		// Convert to ViewModel
		taskViews[i] = task.ToViewModel(assigned_to, created_by, modified_by, *project)
	}

	return taskViews, nil
}

func (s *TaskService) taskToViewModel(task *models.Task) (view *models.TaskView, err error) {
	users := make(map[int64]*models.User, 3)
	var assignedTo *models.User
	assignedTo, ok := users[task.AssignedTo]
	if !ok {
		assignedTo, err = s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			// NOTE: this should already be wrapped
			return nil, err
		}
		users[task.AssignedTo] = assignedTo
	}

	var createdBy *models.User
	createdBy, ok = users[task.CreatedBy]
	if !ok {
		createdBy, err = s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			// NOTE: this should already be wrapped
			return nil, err
		}
		users[task.AssignedTo] = createdBy
	}

	var modifiedBy *models.User
	modifiedBy, ok = users[task.ModifiedBy]
	if !ok {
		modifiedBy, err = s.usersRepo.GetUserById(task.ModifiedBy)
		if err != nil {
			// NOTE: this should already be wrapped
			return nil, err
		}
		users[task.ModifiedBy] = modifiedBy
	}

	project, err := s.projectRepo.GetProject(task.Project)
	if err != nil {
		// NOTE: this should already be wrapped
		return nil, err
	}

	return task.ToViewModel(assignedTo, createdBy, modifiedBy, *project), nil
}

func (s *TaskService) userBelongsToTaskProject(user_id, task_id int64) (err error) {
	task, err := s.tasksRepo.GetItem(task_id)
	if err != nil {
		// NOTE: Take action
		return err
	}
	return s.projectUsersService.UserBelongsToProject(user_id, task.Project)
}
