package dao

import (
	"axiom-blog/global"
	"axiom-blog/global/common"
	"axiom-blog/global/globalInit"
	"axiom-blog/internal/notify/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

/**
  @author: xichencx@163.com
  @date:2021/12/8
  @description: 仅对axiom_blog_notify表进行操作
**/

type Notify model.Notify

func (n Notify) BeforeCreate(tx *gorm.DB) (err error) {
	result := tx.Find(&n)
	if result.RowsAffected != 0 {
		return result.Error
	}
	return
}

func (n Notify) BeforeUpdate(tx *gorm.DB) (err error) {
	result := tx.Find(&n)
	if result.RowsAffected != 0 {
		return result.Error
	}
	return
}

func (n Notify) GetNotifyList(begin, end time.Time) (notify []model.Notify, err error) {
	db := globalInit.Db
	err = db.Where("not(begin_time > ? or end_time < ?) and state = ?", end, begin, global.ONE).Find(&notify).Error
	return
}

func (n Notify) Creat() (id int, err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx.Create(&n).Commit()
		if tx.Error != nil {
			tx.Rollback()
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return n.Id, err
}

func (n Notify) Update(ctx *gin.Context) (err error) {
	tx := globalInit.Transaction()
	err = func(db *gorm.DB) error {
		e := common.ErrDatabase
		tx.Updates(&n).Commit()
		if tx.Error != nil {
			tx.Rollback()
			e.Message = tx.Error.Error()
			return e
		}
		return nil
	}(tx)
	return err
}
