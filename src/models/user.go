package models

import (
	"time"
)

type User struct {
	Id               int64      `json:"id,omitempty"`
	Email            string     `json:"email,omitempty"`
	FullName         string     `json:"full_name,omitempty"`
	SessionId        string     `json:"session_id,omitempty"`
	SessionStartTime *time.Time `json:"session_start_time"`
	SessionValidTill *time.Time `json:"session_valid_till"`
	LastActiveDate   *time.Time `json:"last_active_date"`
	PasswordHash     string     `json:"password_hash,omitempty"`
	IsDeleted        bool       `json:"is_deleted,omitempty"`
	IsEnabled        bool       `json:"is_enabled,omitempty"`
	IsAdmin          bool       `json:"is_admin,omitempty"`
	CreateDate       *time.Time `json:"create_date"`
	AccessGrantedBy  *int64     `json:"access_granted_by,omitempty"`
}

type UserView struct {
	Id          int64      `json:"id,omitempty"`
	Email       string     `json:"email,omitempty"`
	FullName    string     `json:"full_name,omitempty"`
	CreatedDate *time.Time `json:"create_date,omitempty"`
	IsEnabled   bool       `json:"is_enabled,omitempty"`
}

func (u *User) ToViewModel() UserView {
	return UserView{
		Id:          u.Id,
		Email:       u.Email,
		FullName:    u.FullName,
		CreatedDate: u.CreateDate,
		IsEnabled:   u.IsEnabled,
	}
}
