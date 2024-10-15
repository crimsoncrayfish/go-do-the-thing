package models

import (
	"time"
)

type User struct {
	Id               int64     `json:"id,omitempty"`
	Email            string    `json:"email,omitempty"`
	FullName         string    `json:"full_name,omitmepty"`
	SessionId        string    `json:"session_id,omitempty"`
	SessionStartTime time.Time `json:"session_start_time"`
	SessionValidTill time.Time `json:"session_valid_till"`
	LastActiveDate   time.Time `json:"last_active_date"`
	PasswordHash     string    `json:"password_hash,omitempty"`
	IsDeleted        bool      `json:"is_deleted,omitempty"`
	IsAdmin          bool      `json:"is_admin,omitempty"`
	CreateDate       time.Time `json:"create_date"`
}
