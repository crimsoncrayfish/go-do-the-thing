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

func (s TaskService) CreateTask(currentUserId int64, name, description string, due_date *database.SqLiteTime) (int64, error) {
	task := models.Task{
		Name:         name,
		Description:  description,
		DueDate:      due_date,
		AssignedTo:   currentUserId, // TODO: need to update this
		CreatedBy:    currentUserId,
		CreatedDate:  database.SqLiteNow(),
		ModifiedBy:   currentUserId,
		ModifiedDate: database.SqLiteNow(),
		IsDeleted:    false,
	}

	// NOTE: Validate data

	// NOTE: Take action
	id, err := h.repo.InsertItem(task)
	if err != nil {
		h.logger.Error(err, "failed to insert task")
		form.Errors["Task"] = "failed to create task"
		if err := templ_todo.TaskFormContent("Create", form).Render(r.Context(), w); err != nil {
			assert.NoError(err, source, "failed to notify create failure for task")
			// TODO: what should happen if the fetch fails after create
		}
		return
	}
	return 0, nil
}

func (s TaskService) GetTaskView(id, currentUserId int64) (models.TaskView, error) {
	task, err = s.tasksRepo.GetItem(id)
	if err != nil {
		// TODO: what should happen if the fetch fails after create
		assert.NoError(err, source, "failed to get newly inserted task")
		return
	}

	// NOTE: Success zone
	assignedToUser, err := h.usersRepo.GetUserById(task.AssignedTo)
	if ok := h.handleUserIdNotFound(err, task.AssignedTo); !ok {
		assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
		// TODO: what should happen if the fetch fails after create
		return
	}
	var createdBy *models.User
	if task.CreatedBy == task.AssignedTo {
		createdBy = assignedToUser
	} else {
		createdBy, err = h.usersRepo.GetUserById(task.CreatedBy)
		if ok := h.handleUserIdNotFound(err, task.CreatedBy); !ok {
			assert.NoError(err, source, "how does a task with an created by user id of %d even exist?", task.AssignedTo)
			// TODO: what should happen if the fetch fails after create
			return
		}
	}
}
