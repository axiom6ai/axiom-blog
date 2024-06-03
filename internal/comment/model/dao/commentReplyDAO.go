package dao

import (
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/comment/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

/**
  @author: ethan.chen@axiomroup.cn
  @date:2021/10/11
  @description:评论回复表
**/

type CommentReply model.CommentReply

//UpdateCommentReplyByCid
/**
* @Author: ethan.chen@axiomroup.cn
* @Date: 2021/10/11 15:50
* @Description: 更新、删除回复
* @Params: model.CommentReply
* @Return: error
**/

func (c CommentReply) BeforeCreate(tx *gorm.DB) (err error) {
	result := tx.Find(&c)
	if result.RowsAffected != 0 {
		return result.Error
	}
	return
}

func (c CommentReply) BeforeUpdate(tx *gorm.DB) (err error) {
	result := tx.Find(&c)
	if result.RowsAffected != 0 {
		return result.Error
	}
	return
}

func (c CommentReply) UpdateCommentReplyByCid(ctx *gin.Context) (err error) {
	tx := globalInit.Transaction()
	tx.Model(c)
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx.Select("state").Where("cid", c.Cid).Updates(&c)

		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}

		tx.Commit()
		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return
}

func (c CommentReply) CreateCommentReply(ctx *gin.Context) (replyId uint, err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx.Create(&c)
		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}

		tx.Commit()
		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return c.Id, err
}

func (c CommentReply) SelectByState(state uint) (reply map[uint]model.CommentReply) {
	var replies []model.CommentReply
	replies = []model.CommentReply{}
	replyMap := map[uint]model.CommentReply{}
	globalInit.Db.Model(&model.CommentReply{}).Where("state = ?", state).Find(&replies)
	for _, v := range replies {
		replyMap[v.Id] = v
	}
	return replyMap
}

func (c CommentReply) DeleteCommentReplyById(ctx *gin.Context) (err error) {
	tx := globalInit.Transaction()
	tx.Model(&c)
	reply := CommentReply{}
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx = tx.Select("state").Where("id", c.Id).Find(&reply)
		if reply.State == c.State {
			return nil
		}

		tx.Updates(c)

		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}

		tx.Commit()
		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return
}

func (c CommentReply) UpdateCommentReplyStateByCid(id []int, state int) (err error) {
	//c.Id = uint(id)
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx.Model(&c).Where("id in ?", id).Update("state", state)

		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}

		tx.Commit()
		if tx.Error != nil {
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return
}
