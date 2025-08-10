package announcement

import (
	"go-do-the-thing/src/database/repos"
	"go-do-the-thing/src/helpers/slog"
	"go-do-the-thing/src/models"
	"time"
)

type AnnouncementService struct {
	announcementsRepo *repos.RepoContainer
	logger            slog.Logger
}

const serviceSource = "AnnouncementService"

func SetupAnnouncementService(repoContainer *repos.RepoContainer) AnnouncementService {
	return AnnouncementService{
		announcementsRepo: repoContainer,
		logger:            slog.NewLogger(serviceSource),
	}
}

func (s *AnnouncementService) CreateAnnouncement(currentUserId int64, title string, expiresDate *time.Time) (*models.AnnouncementView, error) {
	if expiresDate == nil {
		defaultExpiry := time.Now().AddDate(0, 0, 30)
		expiresDate = &defaultExpiry
	}

	announcement := &models.Announcement{
		Title:       title,
		CreatedBy:   currentUserId,
		ExpiresDate: expiresDate,
		IsActive:    true,
		IsDeleted:   false,
	}

	createdAnnouncement, err := s.announcementsRepo.GetAnnouncementsRepo().CreateAnnouncement(announcement)
	if err != nil {
		s.logger.Error(err, "Failed to create announcement - user_id: %d, title: %s", currentUserId, title)
		return nil, err
	}

	announcements, err := s.announcementsRepo.GetAnnouncementsRepo().GetUserAnnouncements(currentUserId, 1)
	if err != nil || len(announcements) == 0 {
		s.logger.Error(err, "Failed to get created announcement - user_id: %d, announcement_id: %d", currentUserId, createdAnnouncement.Id)
		return nil, err
	}

	s.logger.Info("Created announcement - id: %d, title: %s", createdAnnouncement.Id, title)
	return &announcements[0], nil
}

func (s *AnnouncementService) GetUserAnnouncements(userId int64, limit int) ([]models.AnnouncementView, error) {
	announcements, err := s.announcementsRepo.GetAnnouncementsRepo().GetUserAnnouncements(userId, limit)
	if err != nil {
		s.logger.Error(err, "Failed to get user announcements - user_id: %d", userId)
		return nil, err
	}

	return announcements, nil
}

func (s *AnnouncementService) GetLatestAnnouncementForUser(userId int64) (*models.AnnouncementView, error) {
	announcement, err := s.announcementsRepo.GetAnnouncementsRepo().GetLatestAnnouncementForUser(userId)
	if err != nil {
		s.logger.Error(err, "Failed to get latest announcement - user_id: %d", userId)
		return nil, err
	}

	return announcement, nil
}

func (s *AnnouncementService) GetAllAnnouncements(currentUserId int64) ([]models.AnnouncementAdminView, error) {
	announcements, err := s.announcementsRepo.GetAnnouncementsRepo().GetAllAnnouncements()
	if err != nil {
		s.logger.Error(err, "Failed to get all announcements - user_id: %d", currentUserId)
		return nil, err
	}

	return announcements, nil
}

func (s *AnnouncementService) DismissAnnouncement(userId, announcementId int64) error {
	err := s.announcementsRepo.GetAnnouncementsRepo().DismissAnnouncement(userId, announcementId)
	if err != nil {
		s.logger.Error(err, "Failed to dismiss announcement - user_id: %d, announcement_id: %d", userId, announcementId)
		return err
	}

	s.logger.Info("Dismissed announcement - user_id: %d, announcement_id: %d", userId, announcementId)
	return nil
}

func (s *AnnouncementService) DeleteAnnouncement(currentUserId, announcementId int64) error {
	err := s.announcementsRepo.GetAnnouncementsRepo().DeleteAnnouncement(announcementId)
	if err != nil {
		s.logger.Error(err, "Failed to delete announcement - user_id: %d, announcement_id: %d", currentUserId, announcementId)
		return err
	}

	s.logger.Info("Deleted announcement - user_id: %d, announcement_id: %d", currentUserId, announcementId)
	return nil
}
