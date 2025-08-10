package analytics_repo

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
)

type AnalyticsRepo struct {
	connection database.DatabaseConnection
	logger     slog.Logger
}

const repoSource = "AnalyticsRepo"

func InitRepo(connection database.DatabaseConnection) *AnalyticsRepo {
	logger := slog.NewLogger(repoSource)
	return &AnalyticsRepo{
		connection: connection,
		logger:     logger,
	}
}

func (r *AnalyticsRepo) GetUserStats() (*models.UserStats, error) {
	r.logger.Info("Getting user statistics")

	query := "SELECT * FROM sp_get_user_stats()"
	row := r.connection.QueryRow(query)

	var stats models.UserStats
	err := row.Scan(
		&stats.TotalUsers,
		&stats.ActiveUsers,
		&stats.InactiveUsers,
		&stats.AdminUsers,
	)

	if err != nil {
		r.logger.Error(err, "Failed to get user statistics")
		return nil, err
	}

	return &stats, nil
}

func (r *AnalyticsRepo) GetProjectStats() (*models.ProjectStats, error) {
	r.logger.Info("Getting project statistics")

	query := "SELECT * FROM sp_get_project_stats()"
	row := r.connection.QueryRow(query)

	var stats models.ProjectStats
	err := row.Scan(
		&stats.TotalProjects,
		&stats.ActiveProjects,
		&stats.CompletedProjects,
		&stats.OverdueProjects,
	)

	if err != nil {
		r.logger.Error(err, "Failed to get project statistics")
		return nil, err
	}

	return &stats, nil
}

func (r *AnalyticsRepo) GetTaskStats() (*models.TaskStats, error) {
	r.logger.Info("Getting task statistics")

	query := "SELECT * FROM sp_get_task_stats()"
	row := r.connection.QueryRow(query)

	var stats models.TaskStats
	err := row.Scan(
		&stats.TotalTasks,
		&stats.PendingTasks,
		&stats.InProgressTasks,
		&stats.CompletedTasks,
		&stats.OverdueTasks,
	)

	if err != nil {
		r.logger.Error(err, "Failed to get task statistics")
		return nil, err
	}

	return &stats, nil
}

func (r *AnalyticsRepo) GetRecentActivity() (*models.RecentActivity, error) {
	r.logger.Info("Getting recent activity statistics")

	query := "SELECT * FROM sp_get_recent_activity()"
	row := r.connection.QueryRow(query)

	var activity models.RecentActivity
	err := row.Scan(
		&activity.NewUsersLast30Days,
		&activity.NewProjectsLast30Days,
		&activity.NewTasksLast30Days,
		&activity.CompletedTasksLast30Days,
	)

	if err != nil {
		r.logger.Error(err, "Failed to get recent activity statistics")
		return nil, err
	}

	return &activity, nil
}

func (r *AnalyticsRepo) GetDailyActivity() ([]models.DailyActivity, error) {
	r.logger.Info("Getting daily activity statistics")

	query := "SELECT * FROM sp_get_daily_activity()"
	rows, err := r.connection.Query(query)
	if err != nil {
		r.logger.Error(err, "Failed to query daily activity statistics")
		return nil, err
	}
	defer rows.Close()

	var activities []models.DailyActivity
	for rows.Next() {
		var activity models.DailyActivity
		err := rows.Scan(
			&activity.ActivityDate,
			&activity.NewUsers,
			&activity.NewProjects,
			&activity.NewTasks,
			&activity.CompletedTasks,
		)
		if err != nil {
			r.logger.Error(err, "Failed to scan daily activity row")
			return nil, err
		}
		activities = append(activities, activity)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error(err, "Error iterating daily activity rows")
		return nil, err
	}

	return activities, nil
}
