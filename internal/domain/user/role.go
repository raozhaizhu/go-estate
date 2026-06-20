package role

type Role int16

const (
	// 查询/更新权限: 仅自己; 数据获取范围: 单日查询
	RoleUser Role = iota + 1

	// 查询/更新权限: 仅自己; 数据获取范围: 单日查询, 范围查询
	RoleVip

	// 查询/更新权限: 所有用户; 数据获取范围: 单日查询, 范围查询, 所有查询
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
