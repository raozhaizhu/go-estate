package role

type Role int16

const (
	RoleUser Role = iota + 1
	RoleVip
	RoleAdmin
)

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleVip, RoleAdmin:
		return true
	}
	return false
}
