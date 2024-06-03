package impl

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/comment/model"
	"axiom-blog/internal/comment/model/dao"
	"axiom-blog/internal/comment/qo"
	"axiom-blog/internal/comment/vo"
	"axiom-blog/middleware/jwt"
	"axiom-blog/pkg/commonFunc/articleCommonFunc"
	"axiom-blog/pkg/commonFunc/likeCommonFunc"
	"axiom-blog/pkg/commonFunc/userCommonFunc"
	"axiom-blog/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"reflect"
	"strconv"
)

type Comment struct{}

/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/12 14:02
* @Description: 查询未删除的评论信息
* @Params: cid
* @Return: model.Comment
**/
func (c Comment) commentInfo(cid int) (comment model.Comment) {
	globalInit.Db.Model(&model.Comment{}).
		Where("cid = ? and state = ?", cid, global.ONE).
		First(&comment)
	return
}

/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/12 14:42
* @Description: 查询token信息
* @Params: *gin.Context
* @Return: info *jwt.CustomClaims, err error
**/
func (c Comment) tokenInfo(ctx *gin.Context) (info *jwt.CustomClaims, err error) {
	return jwt.NewJWT().ParseToken(ctx.Request.Header.Get("token"))
}

//List
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/12 17:13
* @Description: 查询文章所有评论及回复
* @Params:
* @Return:
**/
func (c Comment) List(ctx *gin.Context) {
	listQO := qo.ListQO{}
	util.JsonConvert(ctx, &listQO)
	sn, _ := strconv.ParseInt(listQO.Sn, 10, 64)
	listMap := make(map[int]vo.CommentListVO)

	//通过sn查询文章所有评论，生成以floor为key,listVO为value
	var comments []model.Comment
	globalInit.Db.Model(model.Comment{}).Where("sn", sn).Find(&comments)
	if len(comments) == 0 {
		common.SendResponse(ctx, common.OK, listMap)
		return
	}

	userIDList := make([]uuid.UUID, 5)

	for _, v := range comments {
		commentInfo := vo.CommentListVO{}
		_ = copier.Copy(&commentInfo, &v)
		userIDList = append(userIDList, v.UserID)

		//查询该条评论下所有回复
		var reply []model.CommentReply
		globalInit.Db.Where("cid = ? and state = ?", v.Cid, global.ONE).Find(&reply)

		_ = copier.Copy(&commentInfo.ReplyList, &reply)

		for _, rv := range commentInfo.ReplyList {
			userIDList = append(userIDList, rv.UserID)
		}
		listMap[v.Floor] = commentInfo
	}

	//去重
	deduplication := func(s []uuid.UUID) (result []uuid.UUID) {
		m := make(map[uuid.UUID]bool)
		for _, v := range s {
			if !m[v] && v != uuid.Nil {
				result = append(result, v)
				m[v] = true
			}
		}
		return result
	}
	userList := deduplication(userIDList)

	userMap := userCommonFunc.IUser(userCommonFunc.UserCommonFunc{}).FindUser(ctx, userList, "", "")

	//查询登陆用户uid
	tokenInfo, err := c.tokenInfo(ctx)
	var userID uuid.UUID
	if err == nil {
		userID = tokenInfo.ID
	}

	for k, v := range listMap {
		commentInfo := listMap[k]
		commentInfo.NickName = userMap[v.UserID].Nickname
		commentInfo.Avatar = userMap[v.UserID].Avatar

		if userID != uuid.Nil {
			err, zanInfo := likeCommonFunc.LikeCommonFunc{}.CheckUserZanState(ctx, userID, global.ONE, int64(v.Cid))
			if zanInfo.State != global.ONE && err == common.OK {
				commentInfo.ZanState = true
			}
		}

		if v.ReplyList != nil {
			for rk := range commentInfo.ReplyList {
				commentInfo.ReplyList[rk].NickName = userMap[commentInfo.ReplyList[rk].UserID].Nickname
				commentInfo.ReplyList[rk].Avatar = userMap[commentInfo.ReplyList[rk].UserID].Avatar
			}
		}
		listMap[k] = commentInfo
	}

	common.SendResponse(ctx, common.OK, listMap)
}

//Add
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/11 11:29
* @Description: 添加评论
* @Params:
* @Return:
**/
func (c Comment) Add(ctx *gin.Context) {
	var comment dao.Comment
	claims, _ := c.tokenInfo(ctx)
	comment.UserID = claims.ID

	addQO := new(qo.AddCommentQO)
	util.JsonConvert(ctx, addQO)
	sn, _ := strconv.ParseInt(addQO.Sn, 10, 64)
	commentVO := vo.AddCommentVO{}

	if addQO.Content == "" {
		common.SendResponse(ctx, common.ErrParam, commentVO)
		return
	}

	//查询用户是否存在
	user := userCommonFunc.IUser(userCommonFunc.UserCommonFunc{}).FindUser(ctx, []uuid.UUID{comment.UserID}, "", "")
	if userInfo, ok := user[comment.UserID]; !ok || int(userInfo.State) != global.ONE {
		common.SendResponse(ctx, common.ErrUserNotFound, commentVO)
		return
	}

	//查询文章是否存在
	articleMap := articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		FindPublishedArticlesBySn(ctx, []int64{sn})
	if article, ok := articleMap[sn]; !ok || article.State != global.ONE {
		common.SendResponse(ctx, common.ErrArticleNotExisted, commentVO)
		return
	}

	//查询当前文章楼层数
	var floor int
	globalInit.Db.Model(&comment).
		Select("floor").
		Where("sn", sn).
		Order("floor desc").
		First(&floor)

	//插入评论
	//TODO 后续增加审核功能
	comment.State = global.ZERO
	comment.Sn = sn
	comment.Content = addQO.Content
	comment.Floor = floor + global.ONE

	cid, err := comment.CreateComment(ctx)
	if err != nil {
		common.SendResponse(ctx, err, commentVO)
		return
	}

	//查询添加的评论，返回评论Id
	commentVO.CommentId = cid

	//更新文章扩展表评论数
	err = articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		UpdateArticleEx(ctx, sn, false, true, false, true)
	if err != nil {
		common.SendResponse(ctx, err, commentVO)
		return
	}

	common.SendResponse(ctx, common.OK, commentVO)
}

//Delete
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/11 15:10
* @Description: 删除评论（及附属回复）
* @Params: DeleteCommentQO
* @Return:
**/
func (c Comment) Delete(ctx *gin.Context) {
	deleteQO := qo.DeleteCommentQO{}
	util.JsonConvert(ctx, &deleteQO)

	comment := model.Comment{}
	var commentReply []model.CommentReply

	//查询评论状态，已删除/不存在则直接返回
	comment = c.commentInfo(deleteQO.CommentId)
	if reflect.DeepEqual(model.Comment{}, comment) {
		e := common.ErrParam
		e.Message = "comment not exist or was deleted"
		common.SendResponse(ctx, e, "")
		return
	}

	//查询评论是否存在回复，存在则先删除回复
	globalInit.Db.Model(&model.CommentReply{}).
		Where("cid = ? and state = ?", comment.Cid, global.ONE).
		Find(&commentReply)

	if !reflect.DeepEqual([]model.CommentReply{}, commentReply) {
		//删除该Cid下所有回复
		err := dao.CommentReply{Cid: comment.Cid, State: global.THREE}.
			UpdateCommentReplyByCid(ctx)
		if err != nil {
			common.SendResponse(ctx, err, "")
			return
		}
	}

	//更新评论状态为删除
	err := dao.Comment{Cid: comment.Cid, Content: comment.Content, State: global.THREE}.
		UpdateComment(ctx)
	if err != nil {
		common.SendResponse(ctx, err, "")
		return
	}

	//文章扩展表文章评论数更改
	err = articleCommonFunc.IArticle(articleCommonFunc.ArticleCommonFunc{}).
		UpdateArticleEx(ctx, comment.Sn, false, true, false, false)
	if err != nil {
		e := common.ErrDatabase
		e.Message = err.Error()
		common.SendResponse(ctx, e, "")
		return
	}

	//点赞表更改评论点赞状态
	err = likeCommonFunc.LikeCommonFunc{}.
		UpdateZanSate(ctx, comment.Sn, global.ONE)
	if err != nil {
		common.SendResponse(ctx, err, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
}

//AddReply
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/11 15:10
* @Description: 评论回复
* @Params:
* @Return:
**/
func (c Comment) AddReply(ctx *gin.Context) {
	replyQO := qo.AddCommentReplyQO{}
	util.JsonConvert(ctx, &replyQO)
	replyVO := vo.AddCommentReplyVO{}
	token, _ := c.tokenInfo(ctx)

	//查询评论状态（如果非上线状态则不允许进行回复）
	comment := model.Comment{}
	globalInit.Db.Model(model.Comment{}).
		Where("cid = ? and state = ?", replyQO.CommentId, global.ONE).Find(&comment)
	if reflect.DeepEqual(model.Comment{}, comment) {
		common.SendResponse(ctx, common.ErrParam, replyVO)
		return
	}

	//添加回复
	reply := dao.CommentReply{}
	reply.Cid = uint(replyQO.CommentId)
	reply.UserID = token.ID
	reply.Content = replyQO.Content
	//TODO 后续增加审核功能
	reply.State = global.ZERO
	replyId, err := reply.CreateCommentReply(ctx)
	if err != nil {
		common.SendResponse(ctx, err, replyVO)
		return
	}
	common.SendResponse(ctx, common.OK, replyId)
}

//DeleteReply
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/11 15:10
* @Description: 删除回复
* @Params: DeleteCommentReplyQO
* @Return:
**/
func (c Comment) DeleteReply(ctx *gin.Context) {
	deleteQO := qo.DeleteCommentReplyQO{}
	util.JsonConvert(ctx, &deleteQO)

	err := dao.CommentReply{Id: uint(deleteQO.Id), State: global.THREE}.DeleteCommentReplyById(ctx)
	if err != nil {
		e := common.ErrDatabase
		e.Message = err.Error()
		common.SendResponse(ctx, e, "")
		return
	}
	common.SendResponse(ctx, common.OK, "")
}
