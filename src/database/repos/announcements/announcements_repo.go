package announcements

import (
	"go-do-the-thing/src/database"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type AnnouncementsRepo struct {
	db     database.DatabaseConnection
	logger slog.Logger
}

const repoSource = "AnnouncementsRepo"

func InitRepo(connection database.DatabaseConnection) *AnnouncementsRepo {
	return &AnnouncementsRepo{
		db:     connection,
		logger: slog.NewLogger(repoSource),
	}
}

func (r *AnnouncementsRepo) CreateAnnouncement(announcement *models.Announcement) (*models.Announcement, error) {
	query := `
		INSERT INTO announcements (title, created_by, expires_date, is_active, is_deleted)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_date`

	var id int64
	var createdDate time.Time

	row := r.db.QueryRow(query, announcement.Title, announcement.CreatedBy, announcement.ExpiresDate, announcement.IsActive, announcement.IsDeleted)
	err := row.Scan(&id, &createdDate)
	if err != nil {
		r.logger.Error(err, "Failed to create announcement - title: %s, created_by: %d", announcement.Title, announcement.CreatedBy)
		return nil, err
	}

	announcement.Id = id
	announcement.CreatedDate = &createdDate
	r.logger.Info("Created announcement - id: %d, title: %s", id, announcement.Title)
	return announcement, nil
}

func (r *AnnouncementsRepo) GetUserAnnouncements(userId int64, limit int) ([]models.AnnouncementView, error) {
	query := `SELECT id, title, created_by, created_date, expires_date, creator_name FROM sp_get_user_announcements($1, $2)`

	rows, err := r.db.Query(query, userId, limit)
	if err != nil {
		r.logger.Error(err, "Failed to get user announcements - user_id: %d", userId)
		return nil, err
	}
	defer rows.Close()

	var announcements []models.AnnouncementView
	for rows.Next() {
		var announcement models.AnnouncementView
		err := rows.Scan(&announcement.Id, &announcement.Title, &announcement.CreatedBy, &announcement.CreatedDate, &announcement.ExpiresDate, &announcement.CreatorName)
		if err != nil {
			r.logger.Error(err, "Failed to scan announcement row - user_id: %d", userId)
			return nil, err
		}
		announcements = append(announcements, announcement)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error(err, "Error iterating announcement rows - user_id: %d", userId)
		return nil, err
	}

	return announcements, nil
}

func (r *AnnouncementsRepo) GetLatestAnnouncementForUser(userId int64) (*models.AnnouncementView, error) {
	query := `SELECT id, title, created_by, created_date, expires_date, creator_name FROM sp_get_latest_announcement_for_user($1)`

	var announcement models.AnnouncementView
	row := r.db.QueryRow(query, userId)
	err := row.Scan(&announcement.Id, &announcement.Title, &announcement.CreatedBy, &announcement.CreatedDate, &announcement.ExpiresDate, &announcement.CreatorName)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		r.logger.Error(err, "Failed to get latest announcement - user_id: %d", userId)
		return nil, err
	}

	return &announcement, nil
}

func (r *AnnouncementsRepo) GetAllAnnouncements() ([]models.AnnouncementAdminView, error) {
	query := `SELECT id, title, created_by, created_date, expires_date, is_active, is_deleted, creator_name FROM sp_get_all_announcements()`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error(err, "Failed to get all announcements")
		return nil, err
	}
	defer rows.Close()

	var announcements []models.AnnouncementAdminView
	for rows.Next() {
		var announcement models.AnnouncementAdminView
		err := rows.Scan(&announcement.Id, &announcement.Title, &announcement.CreatedBy, &announcement.CreatedDate, &announcement.ExpiresDate, &announcement.IsActive, &announcement.IsDeleted, &announcement.CreatorName)
		if err != nil {
			r.logger.Error(err, "Failed to scan announcement admin row")
			return nil, err
		}
		announcements = append(announcements, announcement)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error(err, "Error iterating announcement admin rows")
		return nil, err
	}

	return announcements, nil
}

func (r *AnnouncementsRepo) DismissAnnouncement(userId, announcementId int64) error {
	query := `
		INSERT INTO user_announcement_dismissals (user_id, announcement_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, announcement_id) DO NOTHING`

	_, err := r.db.Exec(query, userId, announcementId)
	if err != nil {
		r.logger.Error(err, "Failed to dismiss announcement - user_id: %d, announcement_id: %d", userId, announcementId)
		return err
	}

	r.logger.Info("Dismissed announcement - user_id: %d, announcement_id: %d", userId, announcementId)
	return nil
}

func (r *AnnouncementsRepo) DeleteAnnouncement(announcementId int64) error {
	query := `UPDATE announcements SET is_deleted = true WHERE id = $1`

	result, err := r.db.Exec(query, announcementId)
	if err != nil {
		r.logger.Error(err, "Failed to delete announcement - id: %d", announcementId)
		return err
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.logger.Warn("No announcement found to delete - id: %d", announcementId)
		return pgx.ErrNoRows
	}

	r.logger.Info("Deleted announcement - id: %d", announcementId)
	return nil
}
