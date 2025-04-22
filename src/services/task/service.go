package task_service

import (
	"go-do-the-thing/src/database"
	tasks_repo "go-do-the-thing/src/database/repos/tasks"
	users_repo "go-do-the-thing/src/database/repos/users"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type TaskService struct {
	logger    slog.Logger
	tasksRepo tasks_repo.TasksRepo
	usersRepo users_repo.UsersRepo
}

const serviceSource = "TaskService"

func SetupTaskService(tasksRepo tasks_repo.TasksRepo, usersRepo users_repo.UsersRepo) TaskService {
	return TaskService{
		logger:    slog.NewLogger(serviceSource),
		tasksRepo: tasksRepo,
		usersRepo: usersRepo,
	}
}

func (s TaskService) CreateTask(currentUserId, project_id int64, name, description string, due_date *database.SqLiteTime) (int64, error) {
	// TODO: SHOULD I HANDLE PERMISSIONS HERE? May i create a task for this project
	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      due_date,
		AssignedTo:   currentUserId, // TODO: need to update this
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		Project:      project_id,
		IsDeleted:    false,
	}

	// NOTE: Take action
	id, err := s.tasksRepo.InsertItem(task)
	if err != nil {
		s.logger.Error(err, "failed to insert task")
		return 0, err
	}
	return id, nil
}

func (s TaskService) GetTaskView(id, currentUserId int64) (*models.TaskView, error) {
	task, err := s.tasksRepo.GetItem(id)
	if err != nil {
		return nil, err
	}
	// TODO: SHOULD I HANDLE PERMISSIONS HERE?

	return s.taskToViewModel(task)
}

func (s *TaskService) taskToViewModel(task models.Task) (*models.TaskView, error) {
	users := make(map[int64]*models.User, 3)
	assignedTo, ok := users[task.AssignedTo]
	if !ok {
		owner, err := s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[task.AssignedTo] = owner
	}
	createdBy, ok := users[task.CreatedBy]
	if !ok {
		createdBy, err := s.usersRepo.GetUserById(task.AssignedTo)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[task.AssignedTo] = createdBy
	}
	modifiedBy, ok := users[task.ModifiedBy]
	if !ok {
		modifiedBy, err := s.usersRepo.GetUserById(task.ModifiedBy)
		if err != nil {
			// NOTE: this should already be nicely formatted
			return nil, err
		}
		users[task.ModifiedBy] = modifiedBy
	}
	projectView := task.ToViewModel(assignedTo, createdBy, modifiedBy)

	return &projectView, nil
}
