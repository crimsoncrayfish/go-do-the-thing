package constants

const (
	DateTimeFormat   = "2006-01-02 15:04:05"
	DateFormat       = "2006-01-02"
	PrettyDateFormat = "2 January 2006"
)

type ContextKey string

const (
	AuthContext   ContextKey = "auth.middleware.context"
	AuthUserId    ContextKey = "security.middleware.userId"
	AuthUserEmail ContextKey = "security.middleware.userEmail"
	AuthUserName  ContextKey = "security.middleware.userName"
	AuthIsAdmin   ContextKey = "security.middleware.isAdmin"
)
