package model

import (
	"github.com/google/uuid"
	"time"
)

/**
  @author: xichencx@163.com
  @date:2022/6/14
  @description:文件存储地址表
**/

type File struct {
	// 自增ID
	Id int `gorm:"primaryKey; autoIncrement; column:id"`

	//文件名
	Name string `gorm:"column:name"`

	//操作uid
	UserID uuid.UUID `gorm:"column:user_ID"`

	//文件地址
	Address string `gorm:"column:address"`

	//文件类型（1-图片，2-视频，3-文本，4-其他）
	Type int `gorm:"column:type"`

	//文件状态(0-未删除，1-已删除)
	State int `gorm:"column:state"`

	// 创建时间
	CreatedAt time.Time

	// 更新时间
	UpdatedAt time.Time
}
