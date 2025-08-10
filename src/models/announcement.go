package models

import (
	"time"
)

type Announcement struct {
	Id          int64      `json:"id,omitempty"`
	Title       string     `json:"title,omitempty"`
	CreatedBy   int64      `json:"created_by,omitempty"`
	CreatedDate *time.Time `json:"created_date"`
	ExpiresDate *time.Time `json:"expires_date"`
	IsActive    bool       `json:"is_active,omitempty"`
	IsDeleted   bool       `json:"is_deleted,omitempty"`
}

type AnnouncementView struct {
	Id          int64      `json:"id,omitempty"`
	Title       string     `json:"title,omitempty"`
	CreatedBy   int64      `json:"created_by,omitempty"`
	CreatedDate *time.Time `json:"created_date"`
	ExpiresDate *time.Time `json:"expires_date"`
	CreatorName string     `json:"creator_name,omitempty"`
}

type AnnouncementAdminView struct {
	Id          int64      `json:"id,omitempty"`
	Title       string     `json:"title,omitempty"`
	CreatedBy   int64      `json:"created_by,omitempty"`
	CreatedDate *time.Time `json:"created_date"`
	ExpiresDate *time.Time `json:"expires_date"`
	IsActive    bool       `json:"is_active,omitempty"`
	IsDeleted   bool       `json:"is_deleted,omitempty"`
	CreatorName string     `json:"creator_name,omitempty"`
}

type UserAnnouncementDismissal struct {
	Id             int64      `json:"id,omitempty"`
	UserId         int64      `json:"user_id,omitempty"`
	AnnouncementId int64      `json:"announcement_id,omitempty"`
	DismissedDate  *time.Time `json:"dismissed_date"`
}

func (a *Announcement) ToView(creatorName string) AnnouncementView {
	return AnnouncementView{
		Id:          a.Id,
		Title:       a.Title,
		CreatedBy:   a.CreatedBy,
		CreatedDate: a.CreatedDate,
		ExpiresDate: a.ExpiresDate,
		CreatorName: creatorName,
	}
}

func (a *Announcement) ToAdminView(creatorName string) AnnouncementAdminView {
	return AnnouncementAdminView{
		Id:          a.Id,
		Title:       a.Title,
		CreatedBy:   a.CreatedBy,
		CreatedDate: a.CreatedDate,
		ExpiresDate: a.ExpiresDate,
		IsActive:    a.IsActive,
		IsDeleted:   a.IsDeleted,
		CreatorName: creatorName,
	}
}
