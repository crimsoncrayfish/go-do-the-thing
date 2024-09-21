package models

type ActiveScreens struct {
	IsHome     bool
	IsTodoList bool
	IsProjects bool
	IsError    bool
}

type UserDetails struct {
	FullName string
	Nickname string
	Email    string
	IsAdmin  bool
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
		Nickname: name, //TODO: implement nicknames
		Email:    email,
		IsAdmin:  false,
	}
	return n
}
