package form_models

type ItemStatusFilter int

const (
	All ItemStatusFilter = iota
	Completed
	Incomplete
	OverDue
)

func (s ItemStatusFilter) String() string {
	switch s {
	case Completed:
		return "completed"
	case Incomplete:
		return "incomplete"
	case OverDue:
		return "overdue"
	default:
		return "all"
	}
}

func ParseStatusFilter(s string) ItemStatusFilter {
	switch s {
	case "completed":
		return Completed
	case "incomplete":
		return Incomplete
	case "overdue":
		return OverDue
	default:
		return All
	}
}

type TaskSearchForm struct {
	SearchText    string
	ProjectFilter int64
	StatusFilter  ItemStatusFilter
	TagFilters    []int64
}

func NewTaskSearchForm() TaskSearchForm {
	return TaskSearchForm{
		SearchText:    "",
		ProjectFilter: 0,
		StatusFilter:  All,
		TagFilters:    []int64{},
	}
}

// HasActiveFilters returns true if any filter beyond defaults is active
func (f TaskSearchForm) HasActiveFilters() bool {
	return f.ProjectFilter != 0 || f.StatusFilter != All || len(f.TagFilters) > 0
}

// StatusLabel returns a human-readable label for the status filter
func (s ItemStatusFilter) Label() string {
	switch s {
	case Completed:
		return "Completed"
	case Incomplete:
		return "Incomplete"
	case OverDue:
		return "Overdue"
	default:
		return "All"
	}
}
