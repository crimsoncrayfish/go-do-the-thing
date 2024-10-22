package models

import (
	"go-do-the-thing/src/database"
)

type User struct {
	Id               int64                `json:"id,omitempty"`
	Email            string               `json:"email,omitempty"`
	FullName         string               `json:"full_name,omitmepty"`
	SessionId        string               `json:"session_id,omitempty"`
	SessionStartTime *database.SqLiteTime `json:"session_start_time"`
	SessionValidTill *database.SqLiteTime `json:"session_valid_till"`
	LastActiveDate   *database.SqLiteTime `json:"last_active_date"`
	PasswordHash     string               `json:"password_hash,omitempty"`
	IsDeleted        bool                 `json:"is_deleted,omitempty"`
	IsAdmin          bool                 `json:"is_admin,omitempty"`
	CreateDate       *database.SqLiteTime `json:"create_date"`
}

type UserView struct {
	Id       int64  `json:"id,omitempty"`
	Email    string `json:"email,omitempty"`
	FullName string `json:"full_name,omitmepty"`
}
