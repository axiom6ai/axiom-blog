package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	model2 "axiom-blog/internal/article/model"
	"axiom-blog/internal/like/model"
	"axiom-blog/internal/like/model/dao"
	"axiom-blog/internal/like/qo"
	"axiom-blog/middleware/jwt"
	"axiom-blog/pkg/commonFunc/articleCommonFunc"
	"axiom-blog/pkg/commonFunc/commentCommonFunc"
	"axiom-blog/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"reflect"
	"strconv"
)

var (
	zero   = global.ZERO
	zero64 = int64(zero)
)

type Like struct{}

func tokenInfo(ctx *gin.Context) (Info *jwt.CustomClaims, err error) {
	return jwt.NewJWT().ParseToken(ctx.Request.Header.Get("token"))
}

func isLike(ctx *gin.Context, isCancelLike bool) (e error) {
	likeQO := new(qo.LikeQO)
	util.JsonConvert(ctx, likeQO)
	sn, _ := strconv.ParseInt(likeQO.Sn, 10, 64)

	if (sn == zero64 && likeQO.CommentId == zero) ||
		(sn != zero64 && likeQO.CommentId != zero) {
		common.SendResponse(ctx, common.ErrParam, "")
	}
	token, _ := tokenInfo(ctx)

	var err error

	// 点赞/取消点赞文章
	if sn != zero64 {
		if !isCancelLike {
			err = likeArticle(sn, token.ID)
		} else if isCancelLike {
			err = cancelLikeArticle(sn, token.ID)
		}

		if err != nil {
			e := common.ErrDatabase
			e.Message = err.Error()
			return e
		}
	} else if likeQO.CommentId != zero {
		// 点赞/取消点赞评论
		if !isCancelLike {
			err = likeComment(likeQO.CommentId, token.ID)
		} else if isCancelLike {
			err = cancelLikeComment(likeQO.CommentId, token.ID)
		}

		if err != nil {
			e := common.ErrDatabase
			e.Message = err.Error()
			return e
		}
	}
	return common.OK
}

func likeArticle(sn int64, userID uuid.UUID) (err error) {
	//查询文章是否存在,且已上线
	articleMap := articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		FindPublishedArticlesBySn(&gin.Context{}, []int64{sn})
	articleDetail := articleMap[sn]
	if reflect.DeepEqual(model2.Article{}, articleDetail) {
		e := common.ErrParam
		e.Message = "Invalid Param Or Article State is not Published."
		return e
	}

	//点赞表判断是否存在来增加/修改记录
	err = dao.LikeDAO{}.CreatOrUpdate(userID, global.ZERO, sn, false)
	if err != nil && err != common.OK {
		return err
	}

	if err == common.OK {
		return nil
	}

	//文章扩展表更新
	err = articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		UpdateArticleEx(&gin.Context{}, sn, false, false, true, true)
	return
}

func cancelLikeArticle(sn int64, userID uuid.UUID) (err error) {
	//查询文章是否存在,且已上线
	articleMap := articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		FindPublishedArticlesBySn(&gin.Context{}, []int64{sn})
	article := articleMap[sn]
	if reflect.DeepEqual(model2.Article{}, article) {
		e := common.ErrParam
		e.Message = "Invalid Param Or Article State Is Not Published."
		return e
	}

	zanRecord := dao.LikeDAO{}.SelectZan(userID, global.ZERO, sn)
	if reflect.DeepEqual(zanRecord, model.Zan{}) {
		e := common.ErrDatabase
		e.Message = "Like Record Does Not Exist."
		return e
	}

	//点赞表判断是否存在来增加/修改记录
	err = dao.LikeDAO{}.CreatOrUpdate(userID, global.ZERO, sn, true)
	if err != nil && err != common.OK {
		return err
	}

	if err == common.OK {
		return nil
	}

	//文章扩展表更新
	err = articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		UpdateArticleEx(&gin.Context{}, sn, false, false, true, false)
	return
}

func likeComment(commentId int, userID uuid.UUID) (err error) {
	//查询comment是否存在或处于上线状态,并更新评论表点赞数量
	err = commentCommonFunc.IComment(commentCommonFunc.CommentCommonFunc{}).
		UpdateCommentZan(commentId, true)
	if err != nil {
		return
	}

	//点赞表更新记录
	return dao.LikeDAO{}.CreatOrUpdate(userID, global.ONE, int64(commentId), false)
}

func cancelLikeComment(commentId int, userID uuid.UUID) (err error) {
	//查询comment是否存在或处于上线状态,并更新评论表点赞数量
	err = commentCommonFunc.IComment(commentCommonFunc.CommentCommonFunc{}).UpdateCommentZan(commentId, false)
	if err != nil {
		return
	}

	//点赞表更改记录
	return dao.LikeDAO{}.CreatOrUpdate(userID, global.ONE, int64(commentId), true)
}

func (l Like) Like(ctx *gin.Context) {
	//TODO 判断用户是否存在
	common.SendResponse(ctx, isLike(ctx, false), "")
}

func (l Like) CancelLike(ctx *gin.Context) {
	//TODO 判断用户是否存在
	common.SendResponse(ctx, isLike(ctx, true), "")
}

//func (l Like) Update(ctx *gin.Context, objId int64, state int) (err error) {
//	err = dao.LikeDAO{}.Update(ctx, objId, state)
//	if err == nil {
//		return common.OK
//	}
//	e := common.ErrDatabase
//	e.Message = err.Error()
//	return e
//}
