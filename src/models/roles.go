package models

type Role struct {
	Id          RoleEnum `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type RoleEnum int

const (
	BIG_BOSS RoleEnum = iota
	LITTLE_BOSS
	GRUNT
	PLEB
)
