package models

import (
	"fmt"
)

type TagColor string

const (
	ColorRed     TagColor = "red"
	ColorOrange  TagColor = "orange"
	ColorAmber   TagColor = "amber"
	ColorGreen   TagColor = "green"
	ColorCyan    TagColor = "cyan"
	ColorBlue    TagColor = "blue"
	ColorIndigo  TagColor = "indigo"
	ColorPurple  TagColor = "purple"
	ColorFuchsia TagColor = "fuchsia"
	ColorRose    TagColor = "rose"
)

var allTagColors = []TagColor{
	ColorRed,
	ColorOrange,
	ColorAmber,
	ColorGreen,
	ColorCyan,
	ColorBlue,
	ColorIndigo,
	ColorPurple,
	ColorFuchsia,
	ColorRose,
}

var validTagColors map[TagColor]struct{}

func init() {
	validTagColors = make(map[TagColor]struct{})
	for _, color := range allTagColors {
		validTagColors[color] = struct{}{}
	}
}

func ParseTagColor(s string) (TagColor, error) {
	if _, ok := validTagColors[TagColor(s)]; ok {
		return TagColor(s), nil
	}
	return "", fmt.Errorf("invalid color")
}

type Tag struct {
	Id    int64
	Name  string
	Color TagColor
}

type TagView struct {
	Id    int64
	Name  string
	Color TagColor
}

func NewTagView(id int64, name string, color TagColor) TagView {
	return TagView{
		Id:    id,
		Name:  name,
		Color: color,
	}
}

func (t Tag) ToView() *TagView {
	return &TagView{
		t.Id,
		t.Name,
		t.Color,
	}
}

func NewTag(id int64, name string, color TagColor) *TagView {
	return &TagView{
		Id:    id,
		Name:  name,
		Color: color,
	}
}
