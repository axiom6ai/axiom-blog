package authCommonFunc

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/internal/auth"
	"github.com/google/uuid"
)

/**
  @author: xichencx@163.com
  @date:2022/3/30
  @description:
**/

type IAuth interface {
	//AddUserIntoRole 用户添加角色
	AddUserIntoRole(ID uuid.UUID, rName string) (err error)

	//SelectRole 查询角色是否存在
	SelectRole(rName string) bool
}

type AuthCommonFunc struct{}

// AddUserIntoRole 添加用户-角色关联
func (a AuthCommonFunc) AddUserIntoRole(ID uuid.UUID, rName string) (err error) {
	e, err := auth.GetEnforcer()
	if err != nil {
		return err
	}
	//校验角色是否存在
	hasGroup := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+rName)
	if len(hasGroup) == 0 {
		return common.ErrRoleNotExisted
	}

	result, err := e.AddRoleForUser(global.UserPrefix+ID.String(), global.RolePrefix+rName)
	if err != nil {
		return common.ErrDatabase
	}
	if !result {
		return common.ErrUserExistedInRole
	}
	return common.OK
}

func (a AuthCommonFunc) SelectRole(rName string) (has bool, err error) {
	e, err := auth.GetEnforcer()
	if err != nil {
		ee := common.ErrDatabase
		ee.Message = err.Error()
		return false, ee
	}
	//校验角色是否存在
	hasRole := e.GetFilteredNamedGroupingPolicy("g", 1, global.RolePrefix+rName)
	if len(hasRole) == 0 {
		return false, common.ErrRoleNotExisted
	}
	return true, nil
}
