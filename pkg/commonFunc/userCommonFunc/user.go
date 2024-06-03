package userCommonFunc

import (
	"axiom-blog/internal/user/model"
	"axiom-blog/internal/user/model/dao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/**
  @author: ethan.chen@axiomroup.cn
  @date:2021/10/15
  @description:
**/

type IUser interface {
	// FindUser 服务间查询用户信息
	FindUser(ctx *gin.Context, userIDList []uuid.UUID, name string, email string) (users map[uuid.UUID]model.User)

	//UpdateUserAvatar 更新用户头像
	UpdateUserAvatar(ctx *gin.Context, userID uuid.UUID, avatar string) (err error)
}

type UserCommonFunc struct{}

func (c UserCommonFunc) Get() *UserCommonFunc {
	return new(UserCommonFunc)
}

func (c UserCommonFunc) FindUser(ctx *gin.Context, userIDList []uuid.UUID, name string, email string) (users map[uuid.UUID]model.User) {
	findQO := &dao.UserDAO{
		ID:    userIDList,
		Name:  name,
		Email: email,
	}
	userList := findQO.GetUser(ctx)
	users = map[uuid.UUID]model.User{}
	for _, v := range *userList {
		users[v.ID] = v
	}
	return users
}

func (c UserCommonFunc) UpdateUserAvatar(ctx *gin.Context, userID uuid.UUID, avatar string) (err error) {
	updateQO := &model.User{
		ID:     userID,
		Avatar: avatar,
	}

	userDAO := dao.UserDAO{}
	return userDAO.UpdateUserAvatar(ctx, updateQO)
}
