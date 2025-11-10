package adduser

import "bctbackend/database/models"

type RoleOption struct {
	roleId models.RoleID
}

func (option RoleOption) Render() string {
	return option.roleId.Name()
}
