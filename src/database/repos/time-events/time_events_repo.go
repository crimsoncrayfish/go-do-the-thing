// Package timeeventsrepo implements the repository layer for task events (time tracking, etc.)
package timeeventsrepo

import (
	"encoding/json"
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/errors"
	"go-do-the-thing/src/helpers/logger"
	"go-do-the-thing/src/models"
)

type TaskEventsRepo struct {
	database *database.DatabaseConnection
	logger   *logger.Logger
}

func InitRepo(database *database.DatabaseConnection) *TaskEventsRepo {
	log := logger.NewLogger("TaskEventsRepo")
	return &TaskEventsRepo{
		database: database,
		logger:   log,
	}
}

const insertTaskEvent = `SELECT sp_insert_task_event($1, $2, $3, $4)`

// InsertEvent creates a new task event
func (r *TaskEventsRepo) InsertEvent(taskID, userID int64, eventType models.EventType, eventData map[string]any) (int64, error) {
	var id int64

	// Convert event data to JSON
	eventDataJSON, err := json.Marshal(eventData)
	if err != nil {
		eventDataJSON = []byte("{}")
	}

	err = r.database.QueryRow(insertTaskEvent, taskID, userID, string(eventType), eventDataJSON).Scan(&id)
	if err != nil {
		r.logger.Error("failed to insert task event", "error", err, "task_id", taskID, "user_id", userID, "event_type", eventType)
		return 0, errors.NewDbInsertFailed("task_event", err, "", nil)
	}
	r.logger.Info("Task event created", "id", id, "task_id", taskID, "event_type", eventType)
	return id, nil
}

// InsertTimeEvent creates a new time tracking event (convenience method)
func (r *TaskEventsRepo) InsertTimeEvent(taskID, userID int64, eventType models.EventType) (int64, error) {
	return r.InsertEvent(taskID, userID, eventType, nil)
}

const getTaskTimeSpent = `SELECT sp_get_task_time_spent($1)`

// GetTaskTimeSpent returns the total time spent on a task in seconds
func (r *TaskEventsRepo) GetTaskTimeSpent(taskID int64) (int64, error) {
	var seconds int64
	err := r.database.QueryRow(getTaskTimeSpent, taskID).Scan(&seconds)
	if err != nil {
		r.logger.Error("failed to get task time spent", "error", err, "task_id", taskID)
		return 0, errors.NewDbReadFailed("time_spent", err, "", nil)
	}
	return seconds, nil
}

const isTaskInProgress = `SELECT sp_is_task_in_progress($1)`

// IsTaskInProgress checks if a task is currently being tracked
func (r *TaskEventsRepo) IsTaskInProgress(taskID int64) (bool, error) {
	var inProgress bool
	err := r.database.QueryRow(isTaskInProgress, taskID).Scan(&inProgress)
	if err != nil {
		r.logger.Error("failed to check task in progress", "error", err, "task_id", taskID)
		return false, errors.NewDbReadFailed("task_in_progress", err, "", nil)
	}
	return inProgress, nil
}
