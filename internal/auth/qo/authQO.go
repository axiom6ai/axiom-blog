package qo

import "github.com/google/uuid"

type PermissionQO struct {
	PName string `json:"pName" binding:"required"`
	Uri   string `json:"uri" binding:"required"`
}

type AddRoleQO struct {
	RName string `json:"rName" binding:"required"` //role name
}

type GroupAddPermissionQO struct {
	RName string   `json:"rName" binding:"required"` //role name
	PName []string `json:"pName" binding:"required"` //policy name
}

type DeletePermissionFromRoleQO struct {
	RName string   `json:"rName" binding:"required"` //role name
	PName []string `json:"pName" binding:"required"` //policy name
}

type AddUserIntoRoleQO struct {
	UserID uuid.UUID `json:"userID" binding:"required"`
	RName  string    `json:"rName" binding:"required"` //role name
}

type DeletePermissionQO struct {
	PName string `json:"pName" binding:"required"` //policy name
}

type DeleteRoleQO struct {
	RName []string `json:"rName" binding:"required"` //role name
}

type GetUserRolesQO struct {
	UserID []uuid.UUID `json:"userID" binding:"required"`
}

type DeleteUserRoleQO struct {
	UserID uuid.UUID `json:"userID" binding:"required"` //user id
	RName  string    `json:"rName" binding:"required"`  //role name
}

type UserAddRolesInBatches struct {
	UserID uuid.UUID `json:"userID" binding:"required"`
	RName  []string  `json:"rName" binding:"required"`
}

type UserDeleteRolesInBatches struct {
	UserID uuid.UUID `json:"userID" binding:"required"`
	RName  []string  `json:"rName" binding:"required"`
}
