package models

import "fmt"

type Tag struct {
	Id    int64
	Name  string
	Color string
}

type TagView struct {
	Id    int64
	Name  string
	Color string
}

func NewTag(id int64, color string) TagView {
	return TagView{
		Name:  fmt.Sprintf("test-%d", id),
		Id:    id,
		Color: color,
	}
}
