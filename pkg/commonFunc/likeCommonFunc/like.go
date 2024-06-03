package likeCommonFunc

import (
	"axiom-blog/global/common"
	"axiom-blog/internal/like/model"
	"axiom-blog/internal/like/model/dao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"reflect"
)

/**
  @author: ethan.chen@axiomroup.cn
  @date:2021/10/15
  @description:
**/

type ILike interface {
	// UpdateZanSate 服务间更新点赞表state(0未删除，1删除),objId为文章Id或评论Id
	UpdateZanSate(ctx *gin.Context, objId int64, state int) error

	//CheckUserZanState 查询用户是否对文章、评论进行点赞,objId为文章Id或评论Id
	CheckUserZanState(ctx *gin.Context, ID uuid.UUID, objId int64) (err error, isZan bool)
}

type LikeCommonFunc struct{}

func (c LikeCommonFunc) UpdateZanSate(ctx *gin.Context, objId int64, state int) (err error) {
	err = dao.LikeDAO{}.UpdateZanSate(ctx, objId, state)
	if err == nil {
		return common.OK
	}
	e := common.ErrDatabase
	e.Message = err.Error()
	return e
}

func (c LikeCommonFunc) CheckUserZanState(ctx *gin.Context, ID uuid.UUID, objType int, objId int64) (err error, zanInfo model.Zan) {
	zanInfo = dao.LikeDAO{}.SelectZan(ID, objType, objId)
	if reflect.DeepEqual(zanInfo, model.Zan{}) {
		return common.ErrValidation, zanInfo
	}
	return common.OK, zanInfo
}
