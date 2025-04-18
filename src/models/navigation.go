package models

type ActiveScreens struct {
	IsHome     bool
	IsTodoList bool
	IsProjects bool
	IsError    bool
}

type NavBarObject struct {
	ActiveScreens
}

func NewNavbarObject() NavBarObject {
	return NavBarObject{
		ActiveScreens{false, false, false, false},
	}
}
