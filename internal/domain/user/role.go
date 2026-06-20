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

// 定义身份组
var (
	RoleAtLeastUser  = []Role{RoleUser, RoleVip, RoleAdmin}
	RoleAtLeastVip   = []Role{RoleVip, RoleAdmin}
	RoleAtLeastAdmin = []Role{RoleAdmin}
)
