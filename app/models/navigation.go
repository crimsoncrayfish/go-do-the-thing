package models

type ActiveScreens struct {
	IsHome     bool
	IsTodoList bool
	IsProjects bool
	IsError    bool
}

type UserDetails struct {
	FullName string
	Email    string
}

type NavBarObject struct {
	ActiveScreens
	User UserDetails
}

func NewNavbarObject() NavBarObject {
	return NavBarObject{
		ActiveScreens{false, false, false, false},
		UserDetails{FullName: "", Email: ""},
	}
}

func (n NavBarObject) SetUser(name, email string) NavBarObject {
	n.User = UserDetails{
		FullName: name,
		Email:    email,
	}
	return n
}
