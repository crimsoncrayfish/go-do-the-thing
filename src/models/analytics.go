package models

import "time"

type UserStats struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	AdminUsers    int64 `json:"admin_users"`
}

type ProjectStats struct {
	TotalProjects     int64 `json:"total_projects"`
	ActiveProjects    int64 `json:"active_projects"`
	CompletedProjects int64 `json:"completed_projects"`
	OverdueProjects   int64 `json:"overdue_projects"`
}

type TaskStats struct {
	TotalTasks      int64 `json:"total_tasks"`
	PendingTasks    int64 `json:"pending_tasks"`
	InProgressTasks int64 `json:"in_progress_tasks"`
	CompletedTasks  int64 `json:"completed_tasks"`
	OverdueTasks    int64 `json:"overdue_tasks"`
}

type RecentActivity struct {
	NewUsersLast30Days       int64 `json:"new_users_last_30_days"`
	NewProjectsLast30Days    int64 `json:"new_projects_last_30_days"`
	NewTasksLast30Days       int64 `json:"new_tasks_last_30_days"`
	CompletedTasksLast30Days int64 `json:"completed_tasks_last_30_days"`
}

type DailyActivity struct {
	ActivityDate   time.Time `json:"activity_date"`
	NewUsers       int64     `json:"new_users"`
	NewProjects    int64     `json:"new_projects"`
	NewTasks       int64     `json:"new_tasks"`
	CompletedTasks int64     `json:"completed_tasks"`
}

type AnalyticsDashboard struct {
	UserStats      UserStats       `json:"user_stats"`
	ProjectStats   ProjectStats    `json:"project_stats"`
	TaskStats      TaskStats       `json:"task_stats"`
	RecentActivity RecentActivity  `json:"recent_activity"`
	DailyActivity  []DailyActivity `json:"daily_activity"`
}
