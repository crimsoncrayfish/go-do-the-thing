package tasks_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/models"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type TasksRepo struct {
	database database.DatabaseConnection
	logger   *slog.Logger
}

// NOTE: Depends on: [./projects_repo.go, ./users_repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	logger := slog.With("source", "TasksRepo")
	return &TasksRepo{
		database: database,
		logger:   logger,
	}
}

func scanFromRow(row pgx.Row, task *models.Task) error {
	err := row.Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status,
		&task.AssignedTo,
		&task.DueDate,
		&task.CreatedBy,
		&task.CreatedDate,
		&task.ModifiedBy,
		&task.ModifiedDate,
		&task.IsDeleted,
		&task.Project,
		&task.CompleteDate,
	)
	if err != nil {
		return errors.New(errors.ErrDBGenericError, "failed to scan the row: %w", err)
	}
	return nil
}

func scanFromRows(rows pgx.Rows, task *models.Task) error {
	err := rows.Scan(
		&task.Id,
		&task.Name,
		&task.Description,
		&task.Status,
		&task.AssignedTo,
		&task.DueDate,
		&task.CreatedBy,
		&task.CreatedDate,
		&task.ModifiedBy,
		&task.ModifiedDate,
		&task.IsDeleted,
		&task.Project,
		&task.CompleteDate,
	)
	if err != nil {
		return errors.New(errors.ErrDBGenericError, "failed to scan the rows: %w", err)
	}
	return nil
}

const gettasksByAssignedUser = `SELECT * FROM sp_get_tasks_by_user($1)`

func (r *TasksRepo) GetTasksForUser(userId int64) (tasks []*models.Task, err error) {
	r.logger.Debug("GettasksForUser called", "sql", gettasksByAssignedUser, "userId", userId)
	rows, err := r.database.Query(gettasksByAssignedUser, userId)
	if err != nil {
		r.logger.Error("failed to read tasks for user", "error", err, "sql", gettasksByAssignedUser, "userId", userId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read tasks for user: %w", err)
	}
	defer rows.Close()

	tasks = make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}

		err = scanFromRows(rows, task)
		if err != nil {
			// NOTE: Already wrapped
			r.logger.Error("failed to scan row in GettasksForUser", "error", err, "userId", userId)
			return nil, err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GettasksForUser", "error", err, "userId", userId)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	return tasks, nil
}

const getTasksByAssignedUserAndProject = `SELECT * FROM sp_get_tasks_by_user_and_project($1, $2)`

func (r *TasksRepo) GetTasksForUserAndProject(user_id, project_id int64) (tasks []*models.Task, err error) {
	r.logger.Debug("GettasksForUserAndProject called", "sql", getTasksByAssignedUserAndProject, "user_id", user_id, "project_id", project_id)
	rows, err := r.database.Query(getTasksByAssignedUserAndProject, user_id, project_id)
	if err != nil {
		r.logger.Error("failed to read tasks for user and project", "error", err, "sql", getTasksByAssignedUserAndProject, "user_id", user_id, "project_id", project_id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read tasks for user and project: %w", err)
	}
	defer rows.Close()

	tasks = make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}

		err = scanFromRows(rows, task)
		if err != nil {
			r.logger.Error("failed to scan row in GettasksForUserAndProject", "error", err, "user_id", user_id, "project_id", project_id)
			return nil, err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in GettasksForUserAndProject", "error", err, "user_id", user_id, "project_id", project_id)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GettasksForUserAndProject succeeded", "count", len(tasks), "user_id", user_id, "project_id", project_id)
	return tasks, nil
}

const inserttask = `SELECT sp_insert_task($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *TasksRepo) InsertTask(task models.Task) (id int64, err error) {
	r.logger.Debug("Inserttask called", "sql", inserttask, "task_name", task.Name, "project_id", task.Project)
	err = r.database.QueryRow(
		inserttask,
		task.Name,
		task.Description,
		task.Status,
		task.AssignedTo,
		task.DueDate,
		task.CreatedBy,
		time.Now(),
		task.ModifiedBy,
		time.Now(),
		task.Project,
	).Scan(&id)
	if err != nil {
		r.logger.Error("failed to insert task", "error", err, "sql", inserttask, "task_name", task.Name, "project_id", task.Project)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert task: %w", err)
	}
	r.logger.Info("Task created successfully", "id", id, "name", task.Name, "project", task.Project)
	return id, nil
}

const updatetask = `SELECT sp_update_task($1, $2, $3, $4, $5, $6)`

func (r *TasksRepo) UpdateTask(task models.Task) (err error) {
	r.logger.Debug("UpdateTask called", "sql", updatetask, "task_id", task.Id, "task_name", task.Name)
	_, err = r.database.Exec(updatetask, task.Id, task.Name, task.Description, task.AssignedTo, task.DueDate, task.Project)
	if err != nil {
		r.logger.Error("failed to update the task", "error", err, "sql", updatetask, "task_id", task.Id, "task_name", task.Name)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task updated successfully", "id", task.Id, "name", task.Name)
	return nil
}

const updatetaskStatus = `SELECT sp_update_task_status($1, $2, $3, $4, $5)`

func (r *TasksRepo) UpdateTaskStatus(id int64, completeDate *time.Time, status, modifiedBy int64) (err error) {
	r.logger.Debug("UpdateTaskStatus called", "sql", updatetaskStatus, "id", id, "status", status, "completeDate", completeDate, "modifiedBy", modifiedBy)
	_, err = r.database.Exec(updatetaskStatus, id, status, completeDate, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error("failed to update the task status", "error", err, "sql", updatetaskStatus, "id", id, "status", status, "completeDate", completeDate, "modifiedBy", modifiedBy)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task status updated successfully", "id", id, "status", status)
	return nil
}

const deletetask = `SELECT sp_delete_task($1, $2, $3)`

func (r *TasksRepo) DeleteTask(id, modifiedBy int64) (err error) {
	r.logger.Debug("DeleteTask called", "sql", deletetask, "id", id, "modifiedBy", modifiedBy)
	_, err = r.database.Exec(deletetask, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error("failed to delete the task", "error", err, "sql", deletetask, "id", id, "modifiedBy", modifiedBy)
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete the task: %w", err)
	}
	r.logger.Info("Task deleted successfully", "id", id)
	return nil
}

const restoretask = `SELECT sp_restore_task($1, $2, $3)`

func (r *TasksRepo) RestoreTask(id, modifiedBy int64) (err error) {
	r.logger.Debug("Restoretask called", "sql", restoretask, "id", id, "modifiedBy", modifiedBy)
	_, err = r.database.Exec(restoretask, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error("failed to restore the task", "error", err, "sql", restoretask, "id", id, "modifiedBy", modifiedBy)
		return errors.New(errors.ErrDBUpdateFailed, "failed to restore the task: %w", err)
	}
	r.logger.Info("Task restored successfully", "id", id)
	return nil
}

const gettask = `SELECT * FROM sp_get_task($1)`

func (r *TasksRepo) GetTask(id int64) (*models.Task, error) {
	r.logger.Debug("Gettask called", "sql", gettask, "id", id)
	row := r.database.QueryRow(gettask, id)
	temp := &models.Task{}
	err := scanFromRow(row, temp)
	if err != nil {
		r.logger.Error("failed to get the task", "error", err, "sql", gettask, "id", id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get the task: %w", err)
	}
	r.logger.Debug("Gettask succeeded", "id", id)
	return temp, nil
}

const gettasksCount = `SELECT sp_get_tasks_count($1)`

func (r *TasksRepo) GetTasksCount(userId int64) (int64, error) {
	r.logger.Debug("GettasksCount called", "sql", gettasksCount, "userId", userId)
	row := r.database.QueryRow(gettasksCount, userId)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		r.logger.Error("failed to get tasks count", "error", err, "sql", gettasksCount, "userId", userId)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to get tasks count: %w", err)
	}
	r.logger.Debug("GettasksCount succeeded", "count", count, "userId", userId)
	return count, nil
}

const (
	totalTasks     = `SELECT sp_get_total_project_tasks($1)`
	completedTasks = `SELECT sp_get_completed_project_tasks($1)`
)

func (r *TasksRepo) GetProjectTaskCompletion(projectId int64) (int64, int64, error) {
	var total, completed int64
	err := r.database.QueryRow(totalTasks, projectId).Scan(&total)
	if err != nil {
		r.logger.Error("failed to count total tasks for project", "error", err, "projectId", projectId)
		return 0, 0, err
	}
	err = r.database.QueryRow(completedTasks, projectId).Scan(&completed)
	if err != nil {
		r.logger.Error("failed to count completed tasks for project", "error", err, "projectId", projectId)
		return completed, total, err
	}
	return completed, total, nil
}

const searchTasks = `SELECT * FROM sp_get_tasks_by_search($1, $2)`

func (r TasksRepo) Search(user_id int64, search_text string) (tasks []*models.Task, err error) {
	r.logger.Debug("GetTasksBySearch called", "sql", searchTasks, "user_id", user_id)
	rows, err := r.database.Query(searchTasks, user_id, search_text)
	if err != nil {
		r.logger.Error("failed to read tasks for user", "error", err, "sql", searchTasks, "user_id", user_id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to search for tasks: %w", err)
	}
	defer rows.Close()

	tasks = make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}

		err = scanFromRows(rows, task)
		if err != nil {
			// NOTE: Already wrapped
			r.logger.Error("failed to scan row in GettasksForUser", "error", err, "user_id", user_id)
			return nil, err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error("rows.Err() in SearchTasks", "error", err, "user_id", user_id)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("Search succeeded", "count", len(tasks), "user_id", user_id)
	return tasks, nil
}
