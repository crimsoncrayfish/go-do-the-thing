package navigation

type NavBarObject struct {
	IsHome     bool
	IsTodoList bool
	IsProjects bool
	IsError    bool
}

func NewNavbarObject() NavBarObject {
	return NavBarObject{
		false, false, false, false,
	}
}
