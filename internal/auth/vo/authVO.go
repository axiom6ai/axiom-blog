package vo

import "github.com/google/uuid"

/**
  @author: xichencx@163.com
  @date:2022/4/7
  @description:
**/

type UserRolesVO struct {
	UserID    uuid.UUID `json:"userID"`
	UserName  string    `json:"userName"`
	RoleNames []string  `json:"roleNames"`
}

type Roles struct {
	RoleName []string `json:"roleName"`
}
