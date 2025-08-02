package form_models

type ItemStatusFilter int

const (
	All = iota
	Completed
	Incomplete
	OverDue
)

type TaskSearchForm struct {
	SearchText    string
	ProjectFilter int64
	StatusFilter  ItemStatusFilter
}

func NewTaskSearchForm() TaskSearchForm {
	return TaskSearchForm{
		SearchText:    "",
		ProjectFilter: 0,
		StatusFilter:  All,
	}
}
