package users

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

func ParseRole(value string) (Role, error) {
	role := Role(value)
	switch role {
	case RoleUser, RoleModerator, RoleAdmin:
		return role, nil
	default:
		return "", ErrInvalidRole
	}
}

type Status string

const (
	StatusActive  Status = "active"
	StatusBlocked Status = "blocked"
)

func ParseStatus(value string) (Status, error) {
	status := Status(value)
	switch status {
	case StatusActive, StatusBlocked:
		return status, nil
	default:
		return "", ErrInvalidStatus
	}
}
