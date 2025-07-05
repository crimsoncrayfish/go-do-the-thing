package models

type Screen string

const (
	ScreenHome     Screen = "home"
	ScreenTodo     Screen = "todo"
	ScreenProjects Screen = "projects"
	ScreenAdmin    Screen = "admin"
	ScreenError    Screen = "error"
)
