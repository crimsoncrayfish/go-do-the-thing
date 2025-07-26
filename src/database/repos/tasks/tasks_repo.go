package tasks_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type TasksRepo struct {
	database database.DatabaseConnection
	logger   slog.Logger
}

const repoName = "TasksRepo"

// NOTE: Depends on: [./projects_repo.go, ./users_repo.go]
func InitRepo(database database.DatabaseConnection) *TasksRepo {
	return &TasksRepo{database, slog.NewLogger(repoName)}
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
	r.logger.Debug("GettasksForUser called - sql: %s, params: %v", gettasksByAssignedUser, userId)
	rows, err := r.database.Query(gettasksByAssignedUser, userId)
	if err != nil {
		r.logger.Error(err, "failed to read tasks for user - sql: %s, params: %v", gettasksByAssignedUser, userId)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read tasks for user: %w", err)
	}
	defer rows.Close()

	tasks = make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}

		err = scanFromRows(rows, task)
		if err != nil {
			// NOTE: Already wrapped
			r.logger.Error(err, "failed to scan row in GettasksForUser - params: %v", userId)
			return nil, err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GettasksForUser - params: %v", userId)
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GettasksForUser succeeded - count: %d, params: %v", len(tasks), userId)
	return tasks, nil
}

const getTasksByAssignedUserAndProject = `SELECT * FROM sp_get_tasks_by_user_and_project($1, $2)`

func (r *TasksRepo) GetTasksForUserAndProject(user_id, project_id int64) (tasks []*models.Task, err error) {
	r.logger.Debug("GettasksForUserAndProject called - sql: %s, params: %v", getTasksByAssignedUserAndProject, []int64{user_id, project_id})
	rows, err := r.database.Query(getTasksByAssignedUserAndProject, user_id, project_id)
	if err != nil {
		r.logger.Error(err, "failed to read tasks for user and project - sql: %s, params: %v", getTasksByAssignedUserAndProject, []int64{user_id, project_id})
		return nil, errors.New(errors.ErrDBReadFailed, "failed to read tasks for user and project: %w", err)
	}
	defer rows.Close()

	tasks = make([]*models.Task, 0)
	for rows.Next() {
		task := &models.Task{}

		err = scanFromRows(rows, task)
		if err != nil {
			r.logger.Error(err, "failed to scan row in GettasksForUserAndProject - params: %v", []int64{user_id, project_id})
			return nil, err
		}
		tasks = append(tasks, task)
	}
	err = rows.Err()
	if err != nil {
		r.logger.Error(err, "rows.Err() in GettasksForUserAndProject - params: %v", []int64{user_id, project_id})
		return nil, errors.New(errors.ErrDBGenericError, "some error contained in the rows: %w", err)
	}
	r.logger.Debug("GettasksForUserAndProject succeeded - count: %d, params: %v", len(tasks), []int64{user_id, project_id})
	return tasks, nil
}

const inserttask = `SELECT sp_insert_task($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

func (r *TasksRepo) InsertTask(task models.Task) (id int64, err error) {
	r.logger.Debug("Inserttask called - sql: %s, params: %+v", inserttask, task)
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
		r.logger.Error(err, "failed to insert task - sql: %s, params: %+v", inserttask, task)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to insert task: %w", err)
	}
	r.logger.Info("Task created successfully - id: %d, name: %s, project: %d", id, task.Name, task.Project)
	return id, nil
}

const updatetask = `SELECT sp_update_task($1, $2, $3, $4, $5, $6)`

func (r *TasksRepo) UpdateTask(task models.Task) (err error) {
	r.logger.Debug("UpdateTask called - sql: %s, params: %+v", updatetask, task)
	_, err = r.database.Exec(updatetask, task.Id, task.Name, task.Description, task.AssignedTo, task.DueDate, task.Project)
	if err != nil {
		r.logger.Error(err, "failed to update the task - sql: %s, params: %+v", updatetask, task)
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task updated successfully - id: %d, name: %s", task.Id, task.Name)
	return nil
}

const updatetaskStatus = `SELECT sp_update_task_status($1, $2, $3, $4, $5)`

func (r *TasksRepo) UpdateTaskStatus(id int64, completeDate *time.Time, status, modifiedBy int64) (err error) {
	r.logger.Debug("UpdateTaskStatus called - sql: %s, params: %v", updatetaskStatus, []interface{}{id, status, completeDate, modifiedBy, time.Now()})
	_, err = r.database.Exec(updatetaskStatus, id, status, completeDate, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to update the task status - sql: %s, params: %v", updatetaskStatus, []interface{}{id, status, completeDate, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBUpdateFailed, "failed to update the task: %w", err)
	}
	r.logger.Info("Task status updated successfully - id: %d, status: %d", id, status)
	return nil
}

const deletetask = `SELECT sp_delete_task($1, $2, $3)`

func (r *TasksRepo) DeleteTask(id, modifiedBy int64) (err error) {
	r.logger.Debug("DeleteTask called - sql: %s, params: %v", deletetask, []interface{}{id, modifiedBy, time.Now()})
	_, err = r.database.Exec(deletetask, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to delete the task - sql: %s, params: %v", deletetask, []interface{}{id, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBDeleteFailed, "failed to delete the task: %w", err)
	}
	r.logger.Info("Task deleted successfully - id: %d", id)
	return nil
}

const restoretask = `SELECT sp_restore_task($1, $2, $3)`

func (r *TasksRepo) RestoreTask(id, modifiedBy int64) (err error) {
	r.logger.Debug("Restoretask called - sql: %s, params: %v", restoretask, []interface{}{id, modifiedBy, time.Now()})
	_, err = r.database.Exec(restoretask, id, modifiedBy, time.Now())
	if err != nil {
		r.logger.Error(err, "failed to restore the task - sql: %s, params: %v", restoretask, []interface{}{id, modifiedBy, time.Now()})
		return errors.New(errors.ErrDBUpdateFailed, "failed to restore the task: %w", err)
	}
	r.logger.Info("Task restored successfully - id: %d", id)
	return nil
}

const gettask = `SELECT * FROM sp_get_task($1)`

func (r *TasksRepo) GetTask(id int64) (*models.Task, error) {
	r.logger.Debug("Gettask called - sql: %s, params: %v", gettask, id)
	row := r.database.QueryRow(gettask, id)
	temp := &models.Task{}
	err := scanFromRow(row, temp)
	if err != nil {
		r.logger.Error(err, "failed to get the task - sql: %s, params: %v", gettask, id)
		return nil, errors.New(errors.ErrDBReadFailed, "failed to get the task: %w", err)
	}
	r.logger.Debug("Gettask succeeded - id: %d", id)
	return temp, nil
}

const gettasksCount = `SELECT sp_get_tasks_count($1)`

func (r *TasksRepo) GetTasksCount(userId int64) (int64, error) {
	r.logger.Debug("GettasksCount called - sql: %s, params: %v", gettasksCount, userId)
	row := r.database.QueryRow(gettasksCount, userId)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		r.logger.Error(err, "failed to get tasks count - sql: %s, params: %v", gettasksCount, userId)
		return 0, errors.New(errors.ErrDBReadFailed, "failed to get tasks count: %w", err)
	}
	r.logger.Debug("GettasksCount succeeded - count: %d, params: %v", count, userId)
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
		r.logger.Error(err, "failed to count total tasks for project %d", projectId)
		return 0, 0, err
	}
	err = r.database.QueryRow(completedTasks, projectId).Scan(&completed)
	if err != nil {
		r.logger.Error(err, "failed to count completed tasks for project %d", projectId)
		return completed, total, err
	}
	return completed, total, nil
}
